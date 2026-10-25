package agent

import (
	"errors"
	"fmt"
	"github.com/charlzyx/orbit/internal/pirpc"
	"os"
	"path/filepath"
)

// canonicalSessionPath is worker-only. Resolve existing directory aliases even
// when the leaf was just removed, so an in-flight Open still meets a deletion
// reservation/tombstone rather than recreating that file under an alias.
func canonicalSessionPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	ancestor := absolute
	var suffix []string
	for {
		resolved, err := filepath.EvalSymlinks(ancestor)
		if err == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return filepath.Clean(resolved), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", err
		}
		suffix = append(suffix, filepath.Base(ancestor))
		ancestor = parent
	}
}

// FileInfo is captured on workers. SameFile only compares identity fields; no
// filesystem work is performed while the manager mutex is held.
func readSessionFileIdentity(path string) (string, os.FileInfo, error) {
	canonical, err := canonicalSessionPath(path)
	if err != nil {
		return "", nil, err
	}
	info, err := os.Stat(canonical)
	if err != nil && !os.IsNotExist(err) {
		return "", nil, err
	}
	return canonical, info, nil
}
func sameSessionFile(path string, info os.FileInfo, other string, otherInfo os.FileInfo) bool {
	return path == other || info != nil && otherInfo != nil && os.SameFile(info, otherInfo)
}
func reservedSessionFile(path string, info os.FileInfo, reservations map[string]bool, identities map[string]os.FileInfo) bool {
	for p, on := range reservations {
		if on && sameSessionFile(path, info, p, identities[p]) {
			return true
		}
	}
	return false
}

type sessionFileOwner struct {
	entry    *liveSession
	path     string
	live     bool
	retiring map[string]bool
}

func (m *Manager) sessionFileOwnersLocked() map[string]sessionFileOwner {
	owners := make(map[string]sessionFileOwner, len(m.sessions))
	for key, e := range m.sessions {
		o := sessionFileOwner{entry: e, path: e.state.snapshot.ActiveSession.Path, live: key == m.active || e.busy() || e.restartClosing || e.client != nil && e.client.Alive(), retiring: map[string]bool{}}
		for p, on := range e.retiringSessionPaths {
			if on {
				o.retiring[p] = true
			}
		}
		owners[key] = o
	}
	return owners
}
func equalSessionFileOwners(a, b map[string]sessionFileOwner) bool {
	if len(a) != len(b) {
		return false
	}
	for key, x := range a {
		y, ok := b[key]
		if !ok || x.entry != y.entry || x.path != y.path || x.live != y.live || len(x.retiring) != len(y.retiring) {
			return false
		}
		for p := range x.retiring {
			if !y.retiring[p] {
				return false
			}
		}
	}
	return true
}

// reserveSessionDeletion resolves the complete ownership snapshot off-thread,
// then rechecks the exact entries/path sets under the same lock as Open. It
// fails closed on concurrent identity changes; callers may explicitly retry.
func (m *Manager) reserveSessionDeletion(path string) (map[string]*liveSession, error) {
	_, targetInfo, identityErr := readSessionFileIdentity(path)
	if identityErr != nil {
		return nil, identityErr
	}
	if targetInfo == nil {
		return nil, os.ErrNotExist
	}
	m.mu.Lock()
	before := m.sessionFileOwnersLocked()
	m.mu.Unlock()
	matches := map[string]*liveSession{}
	for key, o := range before {
		paths := make(map[string]bool, len(o.retiring)+1)
		if o.path != "" {
			paths[o.path] = true
		}
		for p := range o.retiring {
			paths[p] = true
		}
		for p := range paths {
			canonical, info, err := readSessionFileIdentity(p)
			if err != nil {
				return nil, fmt.Errorf("cannot verify session file ownership: %w", err)
			}
			if sameSessionFile(canonical, info, path, targetInfo) {
				if o.live {
					return nil, errors.New("cannot delete a session with a live runtime; switch away from the task and stop its runtime first")
				}
				matches[key] = o.entry
			}
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, errors.New("session manager is closed")
	}
	if !equalSessionFileOwners(before, m.sessionFileOwnersLocked()) {
		return nil, errors.New("session ownership changed; retry deletion")
	}
	if reservedSessionFile(path, targetInfo, m.deletingSessions, m.deletingSessionInfo) || reservedSessionFile(path, targetInfo, m.metadataSessions, m.metadataSessionInfo) {
		return nil, errors.New("session file operation is in progress; try again when it finishes")
	}
	if m.deletingSessions == nil {
		m.deletingSessions = map[string]bool{}
	}
	m.deletingSessions[path] = true
	if m.deletingSessionInfo == nil {
		m.deletingSessionInfo = map[string]os.FileInfo{}
	}
	m.deletingSessionInfo[path] = targetInfo
	return matches, nil
}

func (m *Manager) bindStartingSessionPath(key string, entry *liveSession, path string) (string, error) {
	canonical, info, err := readSessionFileIdentity(path)
	if err != nil {
		return "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed || m.sessions[key] != entry || entry.stopped {
		return "", errors.New("session start was superseded")
	}
	if info == nil && len(m.deletingSessions) > 0 || reservedSessionFile(canonical, info, m.deletingSessions, m.deletingSessionInfo) {
		return "", errors.New("session deletion superseded this pending open; reopen explicitly after deletion")
	}
	if reservedSessionFile(canonical, info, m.metadataSessions, m.metadataSessionInfo) {
		return "", errors.New("session metadata operation is in progress; retry opening when it finishes")
	}
	// A missing saved path has no inode left to compare (e.g. case aliases on
	// macOS). Conservatively reject only old queued Opens overlapping a delete;
	// a later explicit Open has a newer birth sequence and remains available.
	for deleted, sequence := range m.deletedSessionSequence {
		if sequence >= entry.birthSequence && sequence != 0 && (info == nil || sameSessionFile(canonical, info, deleted, m.deletedSessionInfo[deleted])) {
			return "", errors.New("session deletion superseded this pending open; reopen explicitly after deletion")
		}
	}
	// Preserve the caller's path spelling/reuse identity. Canonicalization is
	// for ownership admission, not a silent rewrite of the Pi session path.
	return path, nil
}

// Resolve aliases before choosing a reusable metadata client. Never silently
// start a second writer because the current runtime used a different spelling.
type metadataSessionTarget struct {
	key   string
	entry *liveSession
	lease *sessionCommandLease
}

func (m *Manager) reserveMetadataSessionTarget(path string, info os.FileInfo) (*metadataSessionTarget, error) {
	m.mu.Lock()
	before := m.sessionFileOwnersLocked()
	m.mu.Unlock()
	matches := map[string]bool{}
	for key, o := range before {
		if o.path != "" {
			canonical, ownerInfo, err := readSessionFileIdentity(o.path)
			if err != nil {
				return nil, err
			}
			if sameSessionFile(canonical, ownerInfo, path, info) {
				matches[key] = true
			}
		}
		if o.live {
			for p := range o.retiring {
				canonical, ownerInfo, err := readSessionFileIdentity(p)
				if err != nil {
					return nil, err
				}
				if sameSessionFile(canonical, ownerInfo, path, info) {
					return nil, errors.New("session identity transition is in progress; retry renaming when it finishes")
				}
			}
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, pirpc.ErrClosed
	}
	if !equalSessionFileOwners(before, m.sessionFileOwnersLocked()) {
		return nil, errors.New("session ownership changed; retry renaming")
	}
	var selected *liveSession
	var selectedKey string
	for key := range matches {
		e := m.sessions[key]
		if e.loading || e.creatingSession || e.restartClosing || e.stopping {
			return nil, errors.New("session identity transition is in progress; retry renaming when it finishes")
		}
		if e.client != nil && e.client.Alive() {
			if selected != nil && selected.client != e.client {
				return nil, errors.New("multiple runtimes own this session file; stop duplicates before renaming")
			}
			selected, selectedKey = e, key
		}
	}
	if selected == nil {
		return nil, nil
	}
	lease, err := m.reserveSessionCommandLocked(selected, pirpc.SetSessionName)
	if err != nil {
		return nil, err
	}
	return &metadataSessionTarget{selectedKey, selected, lease}, nil
}

// Check close to copy/unlink, not only at request admission. This narrows but
// cannot make an external writer/symlink replacement atomic with unlink.
func verifyReservedSessionIdentity(path string, expected os.FileInfo) error {
	current, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if expected == nil || !current.Mode().IsRegular() || !os.SameFile(expected, current) {
		return errors.New("session file identity changed; retry the operation")
	}
	return nil
}
