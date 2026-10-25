package pirpc

import (
	"context"
	"testing"
	"time"
)

// A user-facing deadline notice must not be used as the lifetime context for
// identity-sensitive commands. The existing Call API already retains the wire
// waiter if its context describes process/manager lifetime instead.
func TestRetainedCompletionSurvivesNoticeAndReceivesLateResponse(t *testing.T) {
	c := helper(t, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := c.Call(ctx, NewCommand("pipeline_hold", nil)); done <- err }()
	waitRetainedPending(t, c)
	notice := time.NewTimer(10 * time.Millisecond)
	defer notice.Stop()
	<-notice.C
	select {
	case err := <-done:
		t.Fatalf("deadline notice released wire waiter: %v", err)
	default:
	}
	if c.PendingCount() != 1 {
		t.Fatal("wire registration lost at notice")
	}
	call(t, c, "pipeline_release", nil)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("late response not retained")
	}
	if c.PendingCount() != 0 {
		t.Fatal("completed waiter leaked")
	}
}
func TestRetainedCompletionLifecycleCleanup(t *testing.T) {
	for _, end := range []string{"close", "death", "cancel"} {
		t.Run(end, func(t *testing.T) {
			c := helper(t, nil)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { _, err := c.Call(ctx, NewCommand("hang", nil)); done <- err }()
			waitRetainedPending(t, c)
			switch end {
			case "close":
				_ = c.Close()
			case "death":
				_, _ = c.Call(context.Background(), NewCommand("exit", nil))
			case "cancel":
				cancel()
			}
			select {
			case err := <-done:
				if err == nil {
					t.Fatal("lifecycle end reported success")
				}
			case <-time.After(2 * time.Second):
				t.Fatal("waiter did not terminate")
			}
			if c.PendingCount() != 0 {
				t.Fatal("pending registration leaked")
			}
			if end == "cancel" && !c.Alive() {
				t.Fatal("context cancel must not claim to terminate remote operation")
			}
		})
	}
}
func waitRetainedPending(t *testing.T, c *Client) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if c.PendingCount() == 1 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("call never registered")
}
