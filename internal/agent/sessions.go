package agent

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ScanSessions reads Pi-owned files without changing them. Sort order is last
// message activity, not filesystem mtime (quota bookkeeping must not reorder it).
func ScanSessions(directory string) ([]SessionRow, error) {
	rows := []SessionRow{}
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && path == directory {
				return nil
			}
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || filepath.Ext(path) != ".jsonl" {
			return nil
		}
		row, err := ReadSession(path)
		if err == nil && row.Summary != "" {
			rows = append(rows, row)
		}
		return nil
	})
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Updated.Equal(rows[j].Updated) {
			return rows[i].Path > rows[j].Path
		}
		return rows[i].Updated.After(rows[j].Updated)
	})
	return rows, err
}

func ReadSession(path string) (SessionRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return SessionRow{}, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	row := SessionRow{Path: path}
	header := false
	for {
		line, err := readSessionRecord(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return row, err
		}
		var entry struct {
			Type, ID, Cwd, Name, Timestamp string
			Message                        json.RawMessage
		}
		if json.Unmarshal(line, &entry) != nil {
			continue
		}
		if !header {
			if entry.Type != "session" {
				return row, errors.New("missing Pi session header")
			}
			header = true
			row.ID = entry.ID
			row.Workspace = entry.Cwd
			row.Updated, _ = time.Parse(time.RFC3339Nano, entry.Timestamp)
			continue
		}
		if entry.Type == "session_info" {
			row.Name = strings.TrimSpace(entry.Name)
		}
		if entry.Type == "message" {
			message := decodeMessage(entry.Message, "")
			timestamp := message.Timestamp
			if timestamp.IsZero() {
				timestamp, _ = time.Parse(time.RFC3339Nano, entry.Timestamp)
			}
			if timestamp.After(row.Updated) {
				row.Updated = timestamp
			}
			if row.Summary == "" && message.Role == "user" {
				row.Summary = truncate(message.Text, 160)
			}
		}
	}
	if !header {
		return row, errors.New("empty Pi session")
	}
	if row.Name == "" {
		row.Name = truncate(row.Summary, 80)
	}
	if row.Name == "" {
		row.Name = "New task"
	}
	return row, nil
}

func readSessionRecord(reader *bufio.Reader) ([]byte, error) {
	var line []byte
	for {
		part, err := reader.ReadSlice('\n')
		if len(line)+len(part) > 32<<20 {
			return nil, errors.New("Pi session entry exceeds 32 MiB")
		}
		line = append(line, part...)
		if err == bufio.ErrBufferFull {
			continue
		}
		if err == io.EOF {
			return nil, io.EOF
		}
		return line, err
	}
}

type cachedSession struct {
	info fs.FileInfo
	row  SessionRow
	err  error
}

// SessionIndex makes the polling fallback proportional to changed session files,
// rather than repeatedly reading every historical transcript/image payload.
type SessionIndex struct {
	mu      sync.Mutex
	entries map[string]cachedSession
}

func (index *SessionIndex) Scan(directory string) ([]SessionRow, error) {
	index.mu.Lock()
	defer index.mu.Unlock()
	if index.entries == nil {
		index.entries = map[string]cachedSession{}
	}
	rows := []SessionRow{}
	seen := map[string]bool{}
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, os.ErrNotExist) && path == directory {
				return nil
			}
			return walkErr
		}
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || filepath.Ext(path) != ".jsonl" {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		seen[path] = true
		cached, exists := index.entries[path]
		if !exists || cached.info.Size() != info.Size() || !cached.info.ModTime().Equal(info.ModTime()) || !os.SameFile(cached.info, info) {
			row, readErr := ReadSession(path)
			cached = cachedSession{info: info, row: row, err: readErr}
			index.entries[path] = cached
		}
		if cached.err == nil && cached.row.Summary != "" {
			rows = append(rows, cached.row)
		}
		return nil
	})
	for path := range index.entries {
		if !seen[path] {
			delete(index.entries, path)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Updated.Equal(rows[j].Updated) {
			return rows[i].Path > rows[j].Path
		}
		return rows[i].Updated.After(rows[j].Updated)
	})
	return rows, err
}

// ReadSessionMessages provides the immediate cold-session transcript while Pi
// boots. get_messages replaces it with the authoritative active-branch snapshot.
func ReadSessionMessages(path string) ([]json.RawMessage, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	var messages []json.RawMessage
	for {
		line, err := readSessionRecord(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		var value struct {
			Type    string
			Message json.RawMessage
		}
		if json.Unmarshal(line, &value) == nil && value.Type == "message" && len(value.Message) > 0 {
			messages = append(messages, value.Message)
		}
	}
	return messages, nil
}
