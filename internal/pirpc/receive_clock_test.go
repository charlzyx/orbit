package pirpc

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"
)

func TestEventParseDoesNotInventReceiveTime(t *testing.T) {
	raw := []byte(`{"type":"fixture_event","timestamp":123}`)
	e, err := ParseEvent(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !e.ReceivedAt.IsZero() {
		t.Fatal("pure parse fabricated receive metadata")
	}
	e.ReceivedAt = time.Now()
	encoded, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "ReceivedAt") || string(e.Raw) != string(raw) {
		t.Fatal("receive clock leaked into protocol JSON")
	}
	if localEvent("process_exited", "fixture").ReceivedAt.IsZero() {
		t.Fatal("local terminal event missing creation clock")
	}
}
func TestReaderStampsArrivalBeforeDelayedDrain(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	c := &Client{config: Config{MaxRecordBytes: 4096, MaxEventBytes: 8192, MaxEventCount: 16}}
	done := make(chan struct{})
	go func() { defer close(done); c.readStdout(reader) }()
	waitCount := func(n int) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			c.mu.Lock()
			count := len(c.events)
			c.mu.Unlock()
			if count >= n {
				return
			}
			time.Sleep(time.Millisecond)
		}
		t.Fatal("reader did not enqueue fixture")
	}
	before := time.Now()
	if _, err := io.WriteString(writer, "{\"type\":\"fixture_first\"}\n"); err != nil {
		t.Fatal(err)
	}
	waitCount(1)
	between := time.Now()
	time.Sleep(15 * time.Millisecond)
	if _, err := io.WriteString(writer, "{\"type\":\"fixture_second\"}\n"); err != nil {
		t.Fatal(err)
	}
	waitCount(2)
	after := time.Now()
	time.Sleep(15 * time.Millisecond)
	events := c.DrainEvents()
	if len(events) != 2 || events[0].ReceivedAt.Before(before) || events[0].ReceivedAt.After(between) || events[1].ReceivedAt.Before(between) || events[1].ReceivedAt.After(after) || !events[1].ReceivedAt.After(events[0].ReceivedAt) {
		t.Fatalf("arrival timestamps did not precede drain: %#v", events)
	}
	writer.Close()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("reader did not stop")
	}
}
