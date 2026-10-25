package orbitui

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/egoist/mygo/ui/markdown/parser"
)

// rustTableCells deliberately matches the pinned Rust split_table_row, not
// GFM escaping. The fallback is restricted to rows otherwise losing cells.
func rustTableCells(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	cells := strings.Split(line, "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return cells
}

const raggedTableMaxCells = 16384
const raggedTableMaxBytes = 1 << 20

// Allocation-free shape scan, also used before creating a fallback subtree.
func raggedTableShape(raw string, columns int) (widest, rows int, extraBody bool) {
	widest = columns
	for line := range strings.SplitSeq(strings.TrimSpace(raw), "\n") {
		line = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(line), "|"), "|")
		n := strings.Count(line, "|") + 1
		if n > widest {
			widest = n
		}
		if rows >= 2 && n > columns {
			extraBody = true
		}
		rows++
	}
	return
}

// Check multiplication by division before any matrix/padding allocation.
// Eight bytes per slot bounds separators, row framing and a synthesized delimiter.
func raggedTableWithinBudget(rows, columns, rawBytes int) bool {
	if rows < 1 || columns < 1 || rawBytes < 0 || rawBytes > raggedTableMaxBytes {
		return false
	}
	if columns > raggedTableMaxCells/rows {
		return false
	}
	cells := rows * columns
	return cells <= (raggedTableMaxBytes-rawBytes)/8
}

func raggedTableProjection(raw string, columns int) (string, bool) {
	widest, count, extraBody := raggedTableShape(raw, columns)
	if count < 3 || columns < 1 || !extraBody {
		return "", false
	}
	if !raggedTableWithinBudget(count, widest, len(raw)) {
		return "", true
	}
	var out strings.Builder
	for i, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		row := rustTableCells(line)
		for len(row) < widest {
			value := ""
			if i == 1 {
				value = "---"
			}
			row = append(row, value)
		}
		out.WriteString("| ")
		out.WriteString(strings.Join(row, " | "))
		out.WriteString(" |\n")
	}
	return out.String(), true
}

type raggedTableState struct {
	raw, projected string
	revision       uint64
}
type raggedTableStateKey struct{}

func (a *App) markdownRaggedTable(c *ui.Context, d *parser.Document, i int) bool {
	b := d.Block(i)
	if b.Kind != parser.Table {
		return false
	}
	// Parser End stops at the last retained cell, not the physical row end.
	// Extend only to that same line's end: never consume the following block.
	end := b.End
	if n := strings.IndexByte(d.Source()[end:], '\n'); n >= 0 {
		end += n
	} else {
		end = len(d.Source())
	}
	raw := d.Source()[b.Start:end]
	_, rows, extraBody := raggedTableShape(raw, len(b.Align))
	if rows < 3 || !extraBody {
		return false
	}
	owner := a.inlineImageToken()
	root := ui.Column(c).Key("ragged-table-owner-" + strconv.FormatUint(owner.generation, 10))
	state := ui.Local(root, raggedTableStateKey{}, func() raggedTableState { return raggedTableState{} })
	if state.revision == 0 || state.raw != raw {
		state.raw = raw
		state.projected, _ = raggedTableProjection(raw, len(b.Align))
		state.revision++
	}
	revision := state.revision
	root.Children(func() {
		ui.Column(c).Key("ragged-table-content-" + strconv.FormatUint(revision, 10)).Children(func() {
			if state.projected == "" {
				ui.Text(c, "Table layout limit reached — showing complete source").TextColor(a.Theme.TextMuted).Unselectable()
				ui.Text(c, raw).Font(a.Theme.CodeFont)
				return
			}
			opts := markdown.Options{Theme: a.md.Theme, ShareSelection: true, OnLink: func(url string) {
				if state.revision == revision && a.ownsInlineImage(owner) {
					a.Dispatch(Action{Kind: "OpenLink", Value: url})
				}
			}}
			// Default table cells already paint images as [alt], never load them.
			// No Block/HTML/resource hook can turn cells into executable content.
			markdown.View(c, state.projected, &opts)
		})
	})
	return true
}
