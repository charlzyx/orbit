package workspace

// Review reads immutable private-index snapshots, never stages the user's
// changes. Parsing and expansion are independent of a UI or running agent.
import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const MaxHydratedPatchBytes = 32 * 1024 * 1024
const MaxRenderedDiffLines = 50000
const DefaultExpansionLines = 100

type ReviewSource string

const (
	ReviewLastTurn    ReviewSource = "last_turn"
	ReviewUncommitted ReviewSource = "uncommitted"
	ReviewUnstaged    ReviewSource = "unstaged"
	ReviewStaged      ReviewSource = "staged"
	ReviewCommitted   ReviewSource = "committed"
	ReviewBranch      ReviewSource = "branch"
)

type ReviewRequest struct {
	Source  ReviewSource
	Session string
	Turn    int
}
type ReviewDiff struct {
	Source          ReviewSource
	From, To        string
	Numstat, Patch  string
	CompleteContext bool
}

func (g Git) branchBase(ctx context.Context) (string, error) {
	current, _ := g.text(ctx, "branch", "--show-current")
	for _, name := range []string{"main", "master"} {
		if current == name {
			continue
		}
		if _, e := g.ResolveCommit(ctx, "refs/heads/"+name); e == nil {
			if base, e := g.text(ctx, "merge-base", "HEAD", name); e == nil && base != "" {
				return base, nil
			}
			break
		}
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
	}
	return g.headOrEmpty(ctx)
}
func (g Git) ReviewRange(ctx context.Context, r ReviewRequest) (from, to string, err error) {
	if ok, e := g.IsRepo(ctx); e != nil || !ok {
		if e == nil {
			e = errors.New("not a Git repository")
		}
		return "", "", e
	}
	switch r.Source {
	case ReviewLastTurn:
		if r.Session == "" {
			return "", "", errors.New("no session is open for turn checkpoints")
		}
		if r.Turn < 1 {
			return "", "", errors.New("the first checkpoint is a baseline, not a completed turn")
		}
		for _, ref := range []string{TurnDiffBaseRef(r.Session, r.Turn), TurnStartRef(r.Session, r.Turn), CheckpointRef(r.Session, r.Turn-1)} {
			if from, err = g.ResolveCommit(ctx, ref); err == nil {
				break
			}
			if ctx.Err() != nil {
				return "", "", ctx.Err()
			}
		}
		if err != nil {
			return "", "", fmt.Errorf("turn's starting checkpoint unavailable: %w", err)
		}
		to, err = g.ResolveCommit(ctx, CheckpointRef(r.Session, r.Turn))
		if err != nil {
			return "", "", fmt.Errorf("turn's ending checkpoint unavailable: %w", err)
		}
	case ReviewUncommitted, ReviewStaged:
		from, err = g.headOrEmpty(ctx)
	case ReviewUnstaged:
		from, err = g.IndexTree(ctx)
	case ReviewCommitted, ReviewBranch:
		from, err = g.branchBase(ctx)
	default:
		return "", "", fmt.Errorf("unknown review source %q", r.Source)
	}
	if err != nil || r.Source == ReviewLastTurn {
		return
	}
	switch r.Source {
	case ReviewStaged:
		to, err = g.IndexTree(ctx)
	case ReviewCommitted:
		to, err = g.headOrEmpty(ctx)
	default:
		to, err = g.CaptureWorktreeCommit(ctx)
	}
	return
}
func (g Git) CollectReview(ctx context.Context, r ReviewRequest) (ReviewDiff, error) {
	d := ReviewDiff{Source: r.Source}
	var e error
	d.From, d.To, e = g.ReviewRange(ctx, r)
	if e != nil {
		return d, e
	}
	return g.collectReviewRange(ctx, d, nil)
}

// CollectReviewAt resolves a selected history commit independently of HEAD.
// A root commit is compared with the empty tree; merge commits use parent one.
// Without a ref, source keeps the six-source workspace semantics (LastTurn
// needs CollectReview with a session and turn). Paths are literal repo paths.
func (g Git) CollectReviewAt(ctx context.Context, source ReviewSource, ref, path string) (ReviewDiff, error) {
	d := ReviewDiff{Source: source}
	var err error
	if ref == "" {
		d.From, d.To, err = g.ReviewRange(ctx, ReviewRequest{Source: source})
	} else {
		d.To, err = g.ResolveCommit(ctx, ref)
		if err == nil {
			var parents string
			parents, err = g.text(ctx, "rev-list", "--parents", "-n", "1", d.To)
			if err == nil {
				fields := strings.Fields(parents)
				if len(fields) > 1 {
					d.From = fields[1]
				} else {
					d.From, err = g.emptyTree(ctx)
				}
			}
		}
	}
	if err != nil {
		return d, err
	}
	var paths []string
	if path != "" {
		paths, err = literalPaths([]string{path})
		if err != nil {
			return d, err
		}
	}
	return g.collectReviewRange(ctx, d, paths)
}
func (g Git) collectReviewRange(ctx context.Context, d ReviewDiff, paths []string) (ReviewDiff, error) {
	var e error
	run := func(mode ...string) (string, error) {
		args := []string{"-c", "core.quotePath=false", "diff", "--no-ext-diff", "--no-color", "--no-textconv"}
		args = append(args, mode...)
		args = append(args, "--no-renames", d.From, d.To, "--")
		if len(paths) == 0 {
			args = append(args, ".")
		} else {
			args = append(args, paths...)
		}
		b, e := g.Run(ctx, args...)
		return strings.ToValidUTF8(string(b), "�"), e
	}
	d.Numstat, e = run("--numstat", "-z")
	if e != nil {
		return d, e
	}
	d.Patch, e = run("--unified=2147483647")
	if e != nil {
		return d, e
	}
	d.CompleteContext = len(d.Patch) <= MaxHydratedPatchBytes
	if !d.CompleteContext {
		d.Patch, e = run("--unified=3")
	}
	return d, e
}

type ReviewFile struct {
	Path                 string
	Additions, Deletions int
	Status               string
	DiffLine             int
}
type ReviewGap struct {
	ID       uint64
	Count    int
	Hidden   []ReviewLine
	Position string
}

func (g *ReviewGap) Expandable() bool { return g != nil && g.Count > 0 && g.Count == len(g.Hidden) }

type ReviewLine struct {
	File             int
	OldLine, NewLine int
	Kind, Content    string
	Gap              *ReviewGap
	Section          *ReviewSection
}
type ReviewSection struct {
	Label                       string
	Files, Additions, Deletions int
}
type ReviewSnapshot struct {
	Source               ReviewSource
	Files                []ReviewFile
	Lines                []ReviewLine
	Additions, Deletions int
	Truncated            bool
	CompleteContext      bool
	headers, sections    []int
}

var hunkHeader = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

func ParseReview(d ReviewDiff) ReviewSnapshot {
	s := ReviewSnapshot{Source: d.Source, CompleteContext: d.CompleteContext}
	// NUL framing is necessary for filenames containing tabs or newlines.
	for _, record := range strings.Split(d.Numstat, "\x00") {
		if record == "" {
			continue
		}
		f := strings.SplitN(record, "\t", 3)
		if len(f) != 3 {
			continue
		}
		a, _ := strconv.Atoi(f[0])
		del, _ := strconv.Atoi(f[1])
		status := "modified"
		if f[0] == "-" || f[1] == "-" {
			status = "binary"
		}
		s.Files = append(s.Files, ReviewFile{f[2], a, del, status, -1})
		s.Additions += a
		s.Deletions += del
	}
	var lines []ReviewLine
	file := -1
	old, newLine := 0, 0
	inHunk := false
	var nextGap uint64 = 1
	for _, raw := range strings.Split(strings.TrimSuffix(d.Patch, "\n"), "\n") {
		if strings.HasPrefix(raw, "diff --git ") {
			file++
			old, newLine = 0, 0
			inHunk = false
			if file >= len(s.Files) {
				break
			}
			lines = append(lines, ReviewLine{File: file, Kind: "file_header", Content: s.Files[file].Path})
			continue
		}
		if file < 0 || file >= len(s.Files) {
			continue
		}
		if strings.HasPrefix(raw, "new file mode ") {
			if s.Files[file].Status != "binary" {
				s.Files[file].Status = "added"
			}
		}
		if strings.HasPrefix(raw, "deleted file mode ") {
			if s.Files[file].Status != "binary" {
				s.Files[file].Status = "deleted"
			}
		}
		if m := hunkHeader.FindStringSubmatch(raw); m != nil {
			startOld, _ := strconv.Atoi(m[1])
			startNew, _ := strconv.Atoi(m[3])
			omitted := startOld - old
			if !inHunk {
				omitted = startOld - 1
			}
			if !d.CompleteContext && omitted > 0 {
				n := omitted
				lines = append(lines, ReviewLine{File: file, Kind: "gap", Gap: &ReviewGap{ID: nextGap, Count: n, Position: "between"}})
				nextGap++
			}
			old, newLine = startOld, startNew
			inHunk = true
			lines = append(lines, ReviewLine{File: file, Kind: "hunk_header", Content: raw})
			continue
		}
		l := ReviewLine{File: file, Kind: "meta", Content: raw}
		if inHunk && len(raw) > 0 {
			switch raw[0] {
			case ' ':
				l.Kind = "context"
				l.Content = raw[1:]
				l.OldLine = old
				l.NewLine = newLine
				old++
				newLine++
			case '+':
				l.Kind = "addition"
				l.Content = raw[1:]
				l.NewLine = newLine
				newLine++
			case '-':
				l.Kind = "deletion"
				l.Content = raw[1:]
				l.OldLine = old
				old++
			}
		}
		if raw != "" {
			lines = append(lines, l)
		}
	}
	if d.CompleteContext {
		lines = collapseReview(lines, &nextGap)
	}
	if len(lines) > MaxRenderedDiffLines {
		s.Truncated = true
		lines = lines[:MaxRenderedDiffLines]
	}
	s.Lines = lines
	s.reindex()
	return s
}
func collapseReview(lines []ReviewLine, id *uint64) []ReviewLine {
	var out []ReviewLine
	first, last := map[int]int{}, map[int]int{}
	for i, l := range lines {
		if l.Kind == "addition" || l.Kind == "deletion" {
			if _, ok := first[l.File]; !ok {
				first[l.File] = i
			}
			last[l.File] = i
		}
	}
	for i := 0; i < len(lines); {
		if lines[i].Kind != "context" {
			out = append(out, lines[i])
			i++
			continue
		}
		end := i
		for end < len(lines) && lines[end].Kind == "context" && lines[end].File == lines[i].File {
			end++
		}
		f, changed := first[lines[i].File]
		before, after := changed && f < i, changed && last[lines[i].File] >= end
		left, right := 0, 0
		position := "between"
		if before {
			left = 3
		}
		if after {
			right = 3
		}
		if !before {
			position = "leading"
		}
		if !after {
			position = "trailing"
		}
		if (!before && !after) || end-i-left-right <= 1 {
			out = append(out, lines[i:end]...)
		} else {
			out = append(out, lines[i:i+left]...)
			hidden := append([]ReviewLine(nil), lines[i+left:end-right]...)
			out = append(out, ReviewLine{File: lines[i].File, Kind: "gap", Gap: &ReviewGap{ID: *id, Count: len(hidden), Hidden: hidden, Position: position}})
			*id++
			out = append(out, lines[end-right:end]...)
		}
		i = end
	}
	return out
}
func (s *ReviewSnapshot) reindex() {
	s.headers = nil
	s.sections = nil
	for i := range s.Files {
		s.Files[i].DiffLine = -1
	}
	for i, l := range s.Lines {
		if l.Kind == "file_header" {
			s.headers = append(s.headers, i)
			if l.File >= 0 && l.File < len(s.Files) && s.Files[l.File].DiffLine < 0 {
				s.Files[l.File].DiffLine = i
			}
		}
		if l.Kind == "section_header" {
			s.sections = append(s.sections, i)
		}
	}
}
func around(rows []int, index int) (current, next int, ok bool) {
	i := sort.Search(len(rows), func(i int) bool { return rows[i] > index })
	if i == 0 {
		return -1, -1, false
	}
	next = -1
	if i < len(rows) {
		next = rows[i]
	}
	return rows[i-1], next, true
}
func (s ReviewSnapshot) FileHeadersAround(index int) (int, int, bool) {
	return around(s.headers, index)
}
func (s ReviewSnapshot) SectionHeadersAround(index int) (int, int, bool) {
	return around(s.sections, index)
}
func (s ReviewSnapshot) Grouped() bool { return len(s.sections) > 0 }

// ExpandGap preserves the gap ID and caps visible rows. It returns the number
// of replacement rows, or zero when the compact fallback cannot be expanded.
func (s *ReviewSnapshot) ExpandGap(index int, direction string) int {
	if index < 0 || index >= len(s.Lines) {
		return 0
	}
	line := s.Lines[index]
	g := line.Gap
	if !g.Expandable() {
		return 0
	}
	switch direction {
	case "start", "end", "both", "all":
	default:
		return 0
	}
	available := MaxRenderedDiffLines - (len(s.Lines) - 1)
	if available <= 0 {
		s.Truncated = true
		return 0
	}
	requested := DefaultExpansionLines
	if direction == "both" {
		requested *= 2
	}
	if direction == "all" {
		requested = g.Count
	}
	n := min(requested, g.Count)
	if n < g.Count || n > available {
		n = min(n, available-1)
	}
	if n <= 0 {
		s.Truncated = true
		return 0
	}
	if direction == "all" && n < g.Count {
		s.Truncated = true
	}
	start, end := n, 0
	if direction == "end" || (direction == "all" && g.Position == "leading") {
		start, end = 0, n
	}
	if direction == "both" {
		start = (n + 1) / 2
		end = n - start
	}
	var replacement []ReviewLine
	replacement = append(replacement, g.Hidden[:start]...)
	if n < g.Count {
		remaining := append([]ReviewLine(nil), g.Hidden[start:g.Count-end]...)
		line.Gap = &ReviewGap{g.ID, len(remaining), remaining, g.Position}
		replacement = append(replacement, line)
	}
	replacement = append(replacement, g.Hidden[g.Count-end:]...)
	updated := make([]ReviewLine, 0, len(s.Lines)-1+len(replacement))
	updated = append(updated, s.Lines[:index]...)
	updated = append(updated, replacement...)
	updated = append(updated, s.Lines[index+1:]...)
	s.Lines = updated
	s.reindex()
	return len(replacement)
}

// GroupReviews keeps independent staged/unstaged appearances, including files
// present in both, and offsets file indexes for sticky headers and navigation.
func GroupReviews(staged, unstaged ReviewSnapshot) ReviewSnapshot {
	s := ReviewSnapshot{Source: ReviewUncommitted, CompleteContext: staged.CompleteContext && unstaged.CompleteContext, Truncated: staged.Truncated || unstaged.Truncated}
	for n, part := range []ReviewSnapshot{staged, unstaged} {
		if len(part.Files) == 0 {
			continue
		}
		label := "Staged"
		if n == 1 {
			label = "Changes"
		}
		offset := len(s.Files)
		s.Files = append(s.Files, part.Files...)
		s.Lines = append(s.Lines, ReviewLine{File: -1, Kind: "section_header", Content: label, Section: &ReviewSection{label, len(part.Files), part.Additions, part.Deletions}})
		for _, l := range part.Lines {
			l.File += offset
			if l.Gap != nil {
				g := *l.Gap
				g.Hidden = append([]ReviewLine(nil), g.Hidden...)
				for i := range g.Hidden {
					g.Hidden[i].File += offset
				}
				l.Gap = &g
			}
			s.Lines = append(s.Lines, l)
		}
		s.Additions += part.Additions
		s.Deletions += part.Deletions
	}
	if len(s.Lines) > MaxRenderedDiffLines {
		s.Lines = s.Lines[:MaxRenderedDiffLines]
		s.Truncated = true
	}
	s.reindex()
	return s
}

type ReviewSplitRow struct {
	Left, Right *ReviewLine
	Full        *ReviewLine
}

func SplitReview(lines []ReviewLine) []ReviewSplitRow {
	var out []ReviewSplitRow
	for i := 0; i < len(lines); {
		l := lines[i]
		switch l.Kind {
		case "deletion", "addition":
			var dels, adds []ReviewLine
			f := l.File
			for i < len(lines) && lines[i].File == f && (lines[i].Kind == "deletion" || lines[i].Kind == "addition") {
				if lines[i].Kind == "deletion" {
					dels = append(dels, lines[i])
				} else {
					adds = append(adds, lines[i])
				}
				i++
			}
			for j := 0; j < max(len(dels), len(adds)); j++ {
				r := ReviewSplitRow{}
				if j < len(dels) {
					r.Left = &dels[j]
				}
				if j < len(adds) {
					r.Right = &adds[j]
				}
				out = append(out, r)
			}
		case "context":
			out = append(out, ReviewSplitRow{Left: &lines[i], Right: &lines[i]})
			i++
		default:
			out = append(out, ReviewSplitRow{Full: &lines[i]})
			i++
		}
	}
	return out
}
