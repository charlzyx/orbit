package preferences

import (
	"os"
	"path/filepath"
	"sync"
)

// All Store instances participate: workspaces.json's read/edit/write and
// snapshot revision are serialized together, not merely by a per-Store mutex.
// Other preference files keep their existing locking. External writers are not
// covered by this in-process fence.
var workspaceFiles = struct {
	sync.Mutex
	revision    map[string]uint64
	directories map[string]os.FileInfo
}{revision: map[string]uint64{}, directories: map[string]os.FileInfo{}}

func (s *Store) workspaceFileKey() string {
	root := CanonicalPath(s.RootDir)
	key := filepath.Join(root, "workspaces.json")
	if info, err := os.Stat(root); err == nil {
		for existing, known := range workspaceFiles.directories {
			if existing == key || os.SameFile(info, known) {
				workspaceFiles.directories[existing] = info
				return existing
			}
		}
		workspaceFiles.directories[key] = info
	}
	return key
}
func (s *Store) lockWorkspaceFile(name string) func() {
	if name != "workspaces.json" {
		return func() {}
	}
	workspaceFiles.Lock()
	return workspaceFiles.Unlock
}
