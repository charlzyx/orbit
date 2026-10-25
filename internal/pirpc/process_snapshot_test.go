package pirpc

import (
	"testing"
	"time"
)

func TestProcessSnapshotReportsObservedStartExitAndDetachedCode(t *testing.T) {
	before := time.Now()
	client := helper(t, nil)
	first := client.ProcessSnapshot()
	if first.PID != client.PID() || first.PID <= 0 || first.StartedAt.Before(before) || !first.Alive || first.ExitCode != nil || !first.ExitedAt.IsZero() || first.Executable == "" || first.Launcher == "" {
		t.Fatal(first)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	final := client.ProcessSnapshot()
	if final.Alive || final.ExitCode == nil || *final.ExitCode != client.ExitCode() || final.ExitedAt.Before(first.StartedAt) || !final.StartedAt.Equal(first.StartedAt) {
		t.Fatal(final)
	}
	observed := *final.ExitCode
	*final.ExitCode = 99
	if *client.ProcessSnapshot().ExitCode != observed {
		t.Fatal("exit code aliases client memory")
	}
}
