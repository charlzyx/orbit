package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/agent"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigApplyIdleDoesNotSnapshotRuntime(t *testing.T) {
	x := &Controller{Agent: &agent.Manager{}, stop: make(chan struct{})}
	now := time.Now()
	if n := testing.AllocsPerRun(20, func() { x.tickConfigApply(now) }); n != 0 {
		t.Fatalf("idle tick allocated %g times; no runtime snapshot is needed", n)
	}
	// Nil is an additional access sentinel, not a production initialization path.
	x.Agent = nil
	x.tickConfigApply(now)
}

func TestConfigApplyFutureAndFlightStillObserveRuntime(t *testing.T) {
	now := time.Now()
	for _, flight := range []bool{false, true} {
		x := &Controller{Agent: &agent.Manager{}, stop: make(chan struct{}), configApplyAfter: now.Add(time.Hour)}
		if flight {
			x.configApplyFlight = &configApplyFlight{}
		}
		after := x.configApplyAfter
		if n := testing.AllocsPerRun(5, func() { x.tickConfigApply(now) }); n == 0 {
			t.Fatal("pending/flight incorrectly bypassed runtime observation")
		}
		if x.configApplyAfter != after || flight && x.configApplyFlight == nil {
			t.Fatal("waiting intent lost")
		}
	}
}
func TestConfigApplyStoppedDoesNotObserveRuntime(t *testing.T) {
	stop := make(chan struct{})
	close(stop)
	x := &Controller{stop: stop, configApplyAfter: time.Now(), configApplyFlight: &configApplyFlight{}}
	x.tickConfigApply(time.Now())
	if x.configApplyFlight == nil || x.configApplyAfter.IsZero() {
		t.Fatal("stopped tick changed owner state")
	}
}

func BenchmarkConfigApplyIdleWithHistory(b *testing.B) {
	for _, n := range []int{0, 1000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			home := b.TempDir()
			bin := filepath.Join(home, "synthetic-pi")
			script := fmt.Sprintf(`#!/usr/bin/env python3
import sys,json
messages=[{"id":str(i),"role":"user" if i%%2==0 else "assistant","content":[{"type":"text","text":"Synthetic code discussion. "*40}]} for i in range(%d)]
for line in sys.stdin:
 c=json.loads(line);k=c.get("type");data={}
 if k=="get_state":data={"sessionId":"synthetic","sessionFile":"","isStreaming":False}
 elif k=="get_messages":data={"messages":messages}
 elif k=="get_available_models":data={"models":[]}
 elif k=="get_available_thinking_levels":data={"levels":[]}
 elif k=="get_commands":data={"commands":[]}
 elif k=="get_entries":data={"entries":[]}
 print(json.dumps({"type":"response","id":c.get("id"),"command":k,"success":True,"data":data}),flush=True)
`, n)
			if err := os.WriteFile(bin, []byte(script), 0700); err != nil {
				b.Fatal(err)
			}
			m := agent.NewManager(agent.Options{PiBinary: bin, Home: home, UserHome: home, PiAgentDir: filepath.Join(home, "pi"), SessionDir: filepath.Join(home, "sessions"), DisableExtensions: true, DisablePatches: true, DisableCheckpoints: true, Offline: true})
			b.Cleanup(func() { _ = m.Close() })
			if err := m.Open(home, ""); err != nil {
				b.Fatal(err)
			}
			deadline := time.Now().Add(5 * time.Second)
			for {
				m.Tick()
				s := m.Snapshot()
				if s.Connected && len(s.Messages) == n {
					break
				}
				if time.Now().After(deadline) {
					b.Fatal("synthetic history failed to load", s.Status)
				}
				time.Sleep(time.Millisecond)
			}
			x := &Controller{Agent: m, stop: make(chan struct{})}
			now := time.Now()
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				x.tickConfigApply(now)
			}
		})
	}
}
