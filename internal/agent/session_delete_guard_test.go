package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDeleteSessionPreservesActiveAndParkedRuntime(t *testing.T) {
	for _, active := range []bool{true, false} {
		t.Run(map[bool]string{true: "active", false: "parked"}[active], func(t *testing.T) {
			m := testManager(t, nil)
			path := filepath.Join(t.TempDir(), "keep.jsonl")
			raw := []byte(`{"type":"session","version":3,"id":"kept","cwd":"/workspace","timestamp":"2026-01-01T00:00:00Z"}` + "\n")
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			entry := &liveSession{state: newState("/workspace", path), loading: !active}
			m.mu.Lock()
			m.sessions["kept"] = entry
			if active {
				m.active = "kept"
			}
			m.mu.Unlock()
			if err := m.DeleteSession(path); err != nil {
				t.Fatal(err)
			}
			deadline := time.Now().Add(time.Second)
			for time.Now().Before(deadline) {
				m.mu.Lock()
				failure := m.lastError
				m.mu.Unlock()
				if strings.Contains(failure, "live runtime") {
					break
				}
				time.Sleep(time.Millisecond)
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != string(raw) {
				t.Fatal("live session file changed", err)
			}
			m.mu.Lock()
			preserved := m.sessions["kept"] == entry
			failure := m.lastError
			selection := m.active
			m.mu.Unlock()
			if !preserved || !strings.Contains(failure, "live runtime") || (active && selection != "kept") {
				t.Fatal("runtime or selection changed", preserved, failure, selection)
			}
			if _, err := os.Stat(filepath.Join(m.options.Home, ".orbit-pi", "deleted-sessions")); !os.IsNotExist(err) {
				t.Fatal("blocked deletion wrote a backup", err)
			}
		})
	}
}
func TestOpenCannotRaceReservedSessionDeletion(t *testing.T) {
	m := testManager(t, nil)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "reserved.jsonl")
	m.mu.Lock()
	m.deletingSessions = map[string]bool{path: true}
	m.mu.Unlock()
	if err := m.Open(workspace, path); err == nil || !strings.Contains(err.Error(), "deletion") {
		t.Fatal("open ignored delete reservation", err)
	}
	if m.RuntimeSnapshot().Active != nil {
		t.Fatal("blocked open created runtime")
	}
	m.mu.Lock()
	delete(m.deletingSessions, path)
	m.mu.Unlock()
	if err := m.Open(workspace, path); err != nil {
		t.Fatal("released reservation still blocks", err)
	}
}

func TestDeleteStoppedParkedSessionEvictsOnlyClosedEntry(t *testing.T) {
	m := testManager(t, nil)
	path := filepath.Join(t.TempDir(), "closed.jsonl")
	raw := []byte(`{"type":"session","version":3,"id":"closed","cwd":"/workspace","timestamp":"2026-01-01T00:00:00Z"}` + "\n")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	current := &liveSession{state: newState("/workspace", "/other.jsonl")}
	m.mu.Lock()
	m.sessions["closed"] = &liveSession{stopped: true, state: newState("/workspace", path)}
	m.sessions["current"] = current
	m.active = "current"
	m.mu.Unlock()
	if err := m.DeleteSession(path); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		m.mu.Lock()
		gone := m.sessions["closed"] == nil
		m.mu.Unlock()
		if gone {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("closed file not removed", err)
	}
	m.mu.Lock()
	preserved := m.sessions["current"] == current && m.active == "current"
	gone := m.sessions["closed"] == nil
	m.mu.Unlock()
	if !preserved || !gone {
		t.Fatal("wrong runtime eviction", preserved, gone)
	}
}

func TestDeleteCannotRemovePreviousFileDuringCancelledNewSession(t *testing.T) {
	gate := filepath.Join(t.TempDir(), "release")
	m := testManager(t, func(o *Options) {
		o.EnvOverrides["ORBIT_NEW_HOLD_FILE"] = gate
		o.EnvOverrides["ORBIT_NEW_CANCEL"] = "1"
	})
	workspace := t.TempDir()
	path := filepath.Join(workspace, "old.jsonl")
	raw := []byte(`{"type":"session","version":3,"id":"old","cwd":` + strconv.Quote(workspace) + `,"timestamp":"2026-01-01T00:00:00Z"}` + "\n")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.Path == path })
	if err := m.NewSession(workspace); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		if _, err := os.Stat(gate + ".entered"); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("new_session did not reach held RPC")
		}
		time.Sleep(time.Millisecond)
	}
	if err := m.DeleteSession(path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return strings.Contains(s.ConnectionError, "live runtime") })
	if got, err := os.ReadFile(path); err != nil || string(got) != string(raw) {
		t.Fatal("pending transition file changed", err)
	}
	if err := os.WriteFile(gate, []byte("release"), 0600); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.ActiveSession.Path == path && s.Connected })
	if got, err := os.ReadFile(path); err != nil || string(got) != string(raw) {
		t.Fatal("cancel restored missing file", err)
	}
}
func TestSessionMetadataRenameAndDeleteReservationsExcludeEachOther(t *testing.T) {
	for _, rename := range []bool{false, true} {
		t.Run(map[bool]string{false: "deleteBlocked", true: "renameBlocked"}[rename], func(t *testing.T) {
			m := testManager(t, nil)
			path := filepath.Join(t.TempDir(), "saved.jsonl")
			raw := []byte(`{"type":"session","version":3,"id":"saved","cwd":"/workspace","timestamp":"2026-01-01T00:00:00Z"}` + "\n")
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			m.mu.Lock()
			if rename {
				m.deletingSessions = map[string]bool{path: true}
			} else {
				m.metadataSessions = map[string]bool{path: true}
			}
			m.mu.Unlock()
			var err error
			if rename {
				err = m.RenameSavedSession(path, "new")
			} else {
				err = m.DeleteSession(path)
			}
			if err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool { return strings.Contains(s.ConnectionError, "operation is in progress") })
			if got, err := os.ReadFile(path); err != nil || string(got) != string(raw) {
				t.Fatal("reserved file changed", err)
			}
		})
	}
}

func TestRepeatedRestartCannotDropClosingClientPathOwnership(t *testing.T) {
	m := testManager(t, nil)
	path := filepath.Join(t.TempDir(), "retiring.jsonl")
	entry := &liveSession{state: newState(t.TempDir(), ""), loading: true, restartClosing: true, retiringSessionPaths: map[string]bool{path: true}}
	m.mu.Lock()
	m.sessions["closing"] = entry
	m.active = "closing"
	m.mu.Unlock()
	if err := m.Restart(); err == nil || !strings.Contains(err.Error(), "still closing") {
		t.Fatal("second restart bypassed closing owner", err)
	}
	m.mu.Lock()
	preserved := m.sessions["closing"] == entry && entry.retiringSessionPaths[path]
	m.mu.Unlock()
	if !preserved {
		t.Fatal("closing owner discarded")
	}
}

func TestDeleteSessionDirectoryAliasCannotBypassLiveOwnership(t *testing.T) {
	for _, liveAlias := range []bool{false, true} {
		t.Run(fmt.Sprint(liveAlias), func(t *testing.T) {
			m := testManager(t, nil)
			real := t.TempDir()
			alias := filepath.Join(t.TempDir(), "alias")
			if err := os.Symlink(real, alias); err != nil {
				t.Skip("directory symlink unavailable", err)
			}
			path := filepath.Join(real, "saved.jsonl")
			other := filepath.Join(alias, "saved.jsonl")
			raw := []byte(`{"type":"session","version":3,"id":"alias","cwd":` + strconv.Quote(real) + `,"timestamp":"2026-01-01T00:00:00Z"}` + "\n")
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			live, del := path, other
			if liveAlias {
				live, del = other, path
			}
			if err := m.Open(real, live); err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool { return s.Connected })
			if err := m.DeleteSession(del); err != nil {
				t.Fatal(err)
			}
			m.workers.Wait()
			if got, err := os.ReadFile(path); err != nil || string(got) != string(raw) {
				t.Fatal("alias deletion removed live file", err)
			}
			if !strings.Contains(m.Snapshot().ConnectionError, "live runtime") {
				t.Fatal("alias live owner not detected", m.Snapshot().ConnectionError)
			}
		})
	}
}

func aliasSessionFixture(t *testing.T) (string, string, string) {
	t.Helper()
	root := t.TempDir()
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Skip("directory symlink unavailable", err)
	}
	path := filepath.Join(root, "saved.jsonl")
	other := filepath.Join(alias, "saved.jsonl")
	if err := os.WriteFile(path, []byte(`{"type":"session","version":3,"id":"alias","cwd":`+strconv.Quote(root)+`}`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return root, path, other
}
func TestOpenAliasCannotPassCanonicalDeletionReservation(t *testing.T) {
	m := testManager(t, nil)
	root, path, alias := aliasSessionFixture(t)
	m.mu.Lock()
	m.deletingSessions = map[string]bool{path: true}
	m.mu.Unlock()
	if err := m.Open(root, alias); err != nil {
		t.Fatal("worker admission should resolve alias", err)
	}
	waitFor(t, m, func(s Snapshot) bool { return strings.Contains(s.ConnectionError, "deletion") })
	if p := m.RuntimeSnapshot().Active; p == nil || p.Alive || p.PID != 0 {
		t.Fatal("reserved alias started child", p)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
func TestPendingAliasOpenCannotOutliveDeletionReservation(t *testing.T) {
	m := testManager(t, nil)
	root, path, alias := aliasSessionFixture(t)
	// Hold a registered open before its worker runs, then finish the real unlink.
	m.mu.Lock()
	key, entry := m.newEntryLocked(root, alias, "")
	m.deletingSessions = map[string]bool{path: true}
	m.mu.Unlock()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	m.deletedSessionSequence = map[string]uint64{path: m.sequence}
	delete(m.deletingSessions, path)
	m.mu.Unlock()
	m.start(key, entry, root, alias)
	if !strings.Contains(m.Snapshot().ConnectionError, "deletion") {
		t.Fatal("queued alias open ignored completed deletion", m.Snapshot().ConnectionError)
	}
	if p := m.RuntimeSnapshot().Active; p.Alive || p.PID != 0 {
		t.Fatal("queued open spawned child", p)
	}
	// A later explicit user open is a new intent, not accidentally fenced forever.
	if err := m.Open(root, alias); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
}
func TestMetadataAndDeleteDirectoryAliasesShareReservation(t *testing.T) {
	for _, rename := range []bool{false, true} {
		t.Run(fmt.Sprint(rename), func(t *testing.T) {
			m := testManager(t, nil)
			_, path, alias := aliasSessionFixture(t)
			m.mu.Lock()
			if rename {
				m.deletingSessions = map[string]bool{path: true}
			} else {
				m.metadataSessions = map[string]bool{path: true}
			}
			m.mu.Unlock()
			var err error
			if rename {
				err = m.RenameSavedSession(alias, "changed")
			} else {
				err = m.DeleteSession(alias)
			}
			if err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool { return strings.Contains(s.ConnectionError, "operation is in progress") })
			if _, err := os.Stat(path); err != nil {
				t.Fatal("alias bypassed reservation", err)
			}
		})
	}
}
func TestDeletionOwnershipSnapshotDetectsAllIdentityTransitions(t *testing.T) {
	m := testManager(t, nil)
	m.mu.Lock()
	key, e := m.newEntryLocked(t.TempDir(), "/first.jsonl", "")
	before := m.sessionFileOwnersLocked()
	m.mu.Unlock()
	check := func() {
		t.Helper()
		m.mu.Lock()
		equal := equalSessionFileOwners(before, m.sessionFileOwnersLocked())
		m.mu.Unlock()
		if equal {
			t.Fatal("ownership change not detected")
		}
	}
	m.mu.Lock()
	e.retiringSessionPaths = map[string]bool{"/previous.jsonl": true}
	m.mu.Unlock()
	check()
	m.mu.Lock()
	e.retiringSessionPaths = nil
	e.state.snapshot.ActiveSession.Path = "/next.jsonl"
	m.mu.Unlock()
	check()
	m.mu.Lock()
	e.state.snapshot.ActiveSession.Path = "/first.jsonl"
	delete(m.sessions, key)
	m.newEntryLocked(t.TempDir(), "/first.jsonl", "")
	m.mu.Unlock()
	check()
}

func TestDeleteSessionHardlinkCannotBypassLiveOwnership(t *testing.T) {
	m := testManager(t, nil)
	root, path, _ := aliasSessionFixture(t)
	alias := filepath.Join(root, "hardlink.jsonl")
	if err := os.Link(path, alias); err != nil {
		t.Skip("hardlinks unavailable", err)
	}
	if err := m.Open(root, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.DeleteSession(alias); err != nil {
		t.Fatal(err)
	}
	m.workers.Wait()
	if !strings.Contains(m.Snapshot().ConnectionError, "live runtime") {
		t.Fatal("same inode owner not detected", m.Snapshot().ConnectionError)
	}
	if _, err := os.Stat(alias); err != nil {
		t.Fatal("hardlink removed", err)
	}
}
func TestDeleteSessionCaseVariantProtectsLiveFileWhereSupported(t *testing.T) {
	m := testManager(t, nil)
	root, path, _ := aliasSessionFixture(t)
	alias := filepath.Join(root, "Saved.jsonl")
	original, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	variant, err := os.Stat(alias)
	if err != nil || !os.SameFile(original, variant) {
		t.Skip("filesystem is case-sensitive; Darwin case-insensitive CI must run this case")
	}
	if err := m.Open(root, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.DeleteSession(alias); err != nil {
		t.Fatal(err)
	}
	m.workers.Wait()
	if !strings.Contains(m.Snapshot().ConnectionError, "live runtime") {
		t.Fatal("case alias owner not detected", m.Snapshot().ConnectionError)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("case alias deleted live file", err)
	}
}
func TestSameFileReservationsProtectHardlinkOpenRenameAndDelete(t *testing.T) {
	for _, op := range []string{"open", "rename", "delete"} {
		t.Run(op, func(t *testing.T) {
			m := testManager(t, nil)
			root, path, _ := aliasSessionFixture(t)
			alias := filepath.Join(root, "hardlink.jsonl")
			if err := os.Link(path, alias); err != nil {
				t.Skip("hardlinks unavailable", err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			m.mu.Lock()
			if op == "delete" {
				m.metadataSessions = map[string]bool{path: true}
				m.metadataSessionInfo = map[string]os.FileInfo{path: info}
			} else {
				m.deletingSessions = map[string]bool{path: true}
				m.deletingSessionInfo = map[string]os.FileInfo{path: info}
			}
			m.mu.Unlock()
			switch op {
			case "open":
				err = m.Open(root, alias)
			case "rename":
				err = m.RenameSavedSession(alias, "new")
			case "delete":
				err = m.DeleteSession(alias)
			}
			if err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool { return s.ConnectionError != "" })
			if _, err := os.Stat(alias); err != nil {
				t.Fatal("reservation bypass removed link", err)
			}
			if op == "open" {
				if p := m.RuntimeSnapshot().Active; p.Alive || p.PID != 0 {
					t.Fatal("samefile reservation spawned child", p)
				}
			}
		})
	}
}
func TestMissingLeafOldOpenFailsClosedButExplicitRetryWorks(t *testing.T) {
	m := testManager(t, nil)
	root := t.TempDir()
	path := filepath.Join(root, "MissingCaseAlias.jsonl")
	m.mu.Lock()
	key, e := m.newEntryLocked(root, path, "")
	m.deletedSessionSequence = map[string]uint64{filepath.Join(root, "different-spelling.jsonl"): m.sequence}
	m.mu.Unlock()
	m.start(key, e, root, path)
	if !strings.Contains(m.Snapshot().ConnectionError, "deletion") {
		t.Fatal("ambiguous missing leaf not fenced")
	}
	if err := m.Open(root, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
}
func TestMetadataAliasChoosesExistingClientWithoutSecondWriter(t *testing.T) {
	m := testManager(t, nil)
	root, path, alias := aliasSessionFixture(t)
	if err := m.Open(root, alias); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	m.mu.Lock()
	original := m.sessions[m.active].client
	m.mu.Unlock()
	canonical, info, err := readSessionFileIdentity(path)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := m.reserveMetadataSessionTarget(canonical, info)
	if err != nil || selected == nil || selected.lease.client != original {
		t.Fatal("alias lookup would spawn a second writer", selected != nil && selected.lease.client == original, err)
	}
	m.releaseSessionCommand(selected.entry, selected.lease)
	// If Rename accidentally takes the temporary path, this executable fails.
	m.configuredBinary.Store("/synthetic/nonexistent-pi")
	if err := m.RenameSavedSession(alias, "same-process"); err != nil {
		t.Fatal(err)
	}
	m.workers.Wait()
	if got := m.Snapshot().ConnectionError; got != "" {
		t.Fatal("metadata rename did not reuse live alias", got)
	}
}
func TestReservedTargetIdentityIsRecheckedBeforeMutation(t *testing.T) {
	_, path, _ := aliasSessionFixture(t)
	expected, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	replacement := filepath.Join(filepath.Dir(path), "replacement.jsonl")
	if err := os.WriteFile(replacement, []byte("replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, path); err != nil {
		t.Fatal(err)
	}
	if err := verifyReservedSessionIdentity(path, expected); err == nil {
		t.Fatal("replacement inode accepted under old reservation")
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "replacement" {
		t.Fatal("identity guard changed replacement", err)
	}
}

func TestMissingLeafOpenCannotSlipBetweenUnlinkAndReservationRelease(t *testing.T) {
	m := testManager(t, nil)
	root := t.TempDir()
	missing := filepath.Join(root, "CaseVariant.jsonl")
	m.mu.Lock()
	m.deletingSessions = map[string]bool{filepath.Join(root, "casevariant.jsonl"): true}
	m.mu.Unlock()
	if err := m.Open(root, missing); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return strings.Contains(s.ConnectionError, "deletion") })
	if p := m.RuntimeSnapshot().Active; p.Alive || p.PID != 0 {
		t.Fatal("missing alias started while delete reservation active", p)
	}
}
