package pirpc

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCallOrderedEnqueuesBeforeAwaitAndCorrelatesOutOfOrderReplies(t *testing.T) {
	c := helper(t, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var responses []Response
	c.CallOrdered(ctx, []Command{NewCommand("pipeline_hold", nil), NewCommand("pipeline_release", nil)}, func(i int, r Response, err error) {
		if err != nil {
			t.Errorf("command %d: %v", i, err)
		}
		if i != len(responses) {
			t.Errorf("callback order %d, want %d", i, len(responses))
		}
		responses = append(responses, r)
	})
	if len(responses) != 2 || responses[0].Command != "pipeline_hold" || responses[1].Command != "pipeline_release" {
		t.Fatal("lost response correlation", responses)
	}
	var data struct{ ReceivedAfterHold bool }
	if err := responses[1].DecodeData(&data); err != nil || !data.ReceivedAfterHold {
		t.Fatal("wire order was not hold then release", data, err)
	}
	if c.PendingCount() != 0 {
		t.Fatal("pending calls leaked")
	}
	var wireResponses []string
	for _, e := range c.DrainEvents() {
		if e.Type == "response" {
			var r Response
			_ = e.Decode(&r)
			wireResponses = append(wireResponses, r.Command)
		}
	}
	if strings.Join(wireResponses, ",") != "pipeline_release,pipeline_hold" {
		t.Fatal("fixture did not prove reversed reply order", wireResponses)
	}
}

func TestCallOrderedKeepsIndividualFailuresAndCleansPending(t *testing.T) {
	c := helper(t, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	commands := []Command{NewCommand("fail", nil), NewCommand("mismatch", nil), NewCommand("", nil), NewCommand(GetState, nil)}
	results := 0
	c.CallOrdered(ctx, commands, func(i int, r Response, err error) {
		results++
		if i < 3 && err == nil {
			t.Errorf("command %d lost its independent error", i)
		}
		if i == 3 && (err != nil || r.Command != GetState) {
			t.Errorf("earlier rejection suppressed later command: %v %v", r, err)
		}
	})
	if results != len(commands) || c.PendingCount() != 0 {
		t.Fatal(results, c.PendingCount())
	}
}

func TestCallOrderedCancellationRetainsCompletedLaterResponse(t *testing.T) {
	c := helper(t, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan struct{})
	var gotErrors [2]error
	var gotResponses [2]Response
	go func() {
		defer close(finished)
		c.CallOrdered(ctx, []Command{NewCommand("hang", nil), NewCommand(GetState, nil)}, func(i int, r Response, err error) { gotErrors[i], gotResponses[i] = err, r })
	}()
	// Observe the actual later reply and then its removal from pending under
	// the reader's lock. Event enqueue alone precedes response correlation.
	deadline := time.Now().Add(2 * time.Second)
	responseID := ""
	completed := false
	for !completed && time.Now().Before(deadline) {
		for _, e := range c.DrainEvents() {
			if e.Type == "response" {
				var r Response
				_ = e.Decode(&r)
				if r.Command == GetState {
					responseID = r.ID
				}
			}
		}
		if responseID != "" {
			c.mu.Lock()
			_, pending := c.pending[responseID]
			completed = !pending
			c.mu.Unlock()
		}
		if !completed {
			time.Sleep(time.Millisecond)
		}
	}
	if !completed {
		t.Fatal("later response never completed")
	}
	cancel()
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("cancel did not finish pipeline")
	}
	if !errors.Is(gotErrors[0], context.Canceled) || gotErrors[1] != nil || gotResponses[1].Command != GetState {
		t.Fatal(gotErrors, gotResponses)
	}
	if c.PendingCount() != 0 {
		t.Fatal("cancel leaked pending calls")
	}
}

func TestCallOrderedReleasesAllRegistrationsIfReceiverPanics(t *testing.T) {
	c := helper(t, nil)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("receiver did not panic")
			}
		}()
		c.CallOrdered(context.Background(), []Command{NewCommand(GetState, nil), NewCommand("hang", nil)}, func(int, Response, error) { panic("receiver stopped") })
	}()
	if c.PendingCount() != 0 {
		t.Fatal("panic leaked unawaited calls")
	}
}
