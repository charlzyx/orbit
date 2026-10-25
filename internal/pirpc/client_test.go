package pirpc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestHelperProcess is a real subprocess fixture, not a mocked io.Reader.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("ORBIT_RPC_HELPER") != "1" {
		return
	}
	if os.Getenv("ORBIT_RPC_NO_READ") == "1" {
		time.Sleep(time.Hour)
		os.Exit(0)
	}
	var mu sync.Mutex
	write := func(value any) {
		wire, _ := json.Marshal(value)
		mu.Lock()
		fmt.Fprintln(os.Stdout, string(wire))
		mu.Unlock()
	}
	input := bufio.NewReader(os.Stdin)
	for {
		line, err := input.ReadBytes('\n')
		if err != nil {
			break
		}
		var command map[string]any
		if json.Unmarshal(line, &command) != nil {
			continue
		}
		kind, _ := command["type"].(string)
		id, _ := command["id"].(string)
		if isOneWay(kind) {
			write(map[string]any{"type": "ui_received", "id": id, "record": command})
			continue
		}
		response := map[string]any{"type": "response", "id": id, "command": kind, "success": true, "data": command}
		switch kind {
		case "hang":
			continue
		case "exit":
			os.Exit(7)
		case "fail":
			response["success"] = false
			response["error"] = "intentional rejection"
		case "mismatch":
			response["command"] = "different"
		case "delayed":
			delay, _ := command["delay"].(float64)
			go func(r any, d float64) { time.Sleep(time.Duration(d) * time.Millisecond); write(r) }(response, delay)
			continue
		case "inspect":
			cwd, _ := os.Getwd()
			response["data"] = map[string]any{"args": os.Args, "cwd": cwd, "marker": os.Getenv("ORBIT_TEST_MARKER"), "skip": os.Getenv("PI_SKIP_VERSION_CHECK"), "offline": os.Getenv("PI_OFFLINE"), "secret": os.Getenv("ORBIT_TEST_SECRET")}
		case "malformed":
			mu.Lock()
			fmt.Fprint(os.Stdout, "not-json\n[]\n{\"type\":\"future_event\",\"novel\":{\"n\":42}}\n")
			mu.Unlock()
		case "large":
			response["data"] = map[string]any{"text": strings.Repeat("λ", 100000) + "\u2028paragraph\u2029end"}
			wire, _ := json.Marshal(response)
			wire = []byte(strings.ReplaceAll(strings.ReplaceAll(string(wire), "\\u2028", "\u2028"), "\\u2029", "\u2029"))
			mu.Lock()
			fmt.Fprint(os.Stdout, string(wire)+"\r\n")
			mu.Unlock()
			continue
		case "stderr":
			for i := 0; i < 300; i++ {
				fmt.Fprintf(os.Stderr, "line-%d:%s\n", i, strings.Repeat("x", 4096))
			}
		case "flood":
			for i := 0; i < 100; i++ {
				write(map[string]any{"type": "future_event", "data": strings.Repeat("x", 2048)})
			}
		}
		write(response)
	}
	if os.Getenv("ORBIT_RPC_IGNORE_EOF") == "1" {
		time.Sleep(time.Hour)
	}
	os.Exit(0)
}

func helper(t *testing.T, modify func(*Config)) *Client {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	config := Config{Executable: executable, PrefixArgs: []string{"-test.run=^TestHelperProcess$", "--"}, WorkspaceDir: t.TempDir(), Environment: []string{}, EnvOverrides: map[string]string{"ORBIT_RPC_HELPER": "1"}, ShutdownTimeout: 250 * time.Millisecond}
	if modify != nil {
		modify(&config)
	}
	client, err := Start(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}
func call(t *testing.T, c *Client, kind string, params map[string]any) Response {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := c.Call(ctx, NewCommand(kind, params))
	if err != nil {
		t.Fatal(err)
	}
	return response
}
func awaitDone(t *testing.T, c *Client) {
	t.Helper()
	select {
	case <-c.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("process did not exit")
	}
}

func TestCommandWireNames(t *testing.T) {
	commands := []string{Prompt, Steer, FollowUp, Abort, ClearQueue, NewSession, SwitchSession, GetState, GetMessages, SetSessionName, GetSessionStats, GetForkMessages, GetEntries, Fork, Clone, GetTree, GetLastAssistantText, GetAvailableModels, SetModel, CycleModel, GetAvailableThinkingLevels, SetThinkingLevel, CycleThinkingLevel, SetSteeringMode, SetFollowUpMode, Compact, SetAutoCompaction, SetAutoRetry, AbortRetry, Bash, AbortBash, ExportHTML, GetCommands, AuthList, AuthStatus, AuthLogin, AuthLogout, AuthCancel, QuotaList}
	for _, kind := range commands {
		command := NewCommand(kind, map[string]any{"futureField": "kept"})
		command.ID = "explicit"
		raw, err := json.Marshal(command)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]any
		if err = json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded["type"] != kind || decoded["id"] != "explicit" || decoded["futureField"] != "kept" {
			t.Fatal(string(raw))
		}
	}
	for _, reserved := range []string{"type", "id"} {
		if _, err := json.Marshal(NewCommand(GetState, map[string]any{reserved: "bad"})); err == nil {
			t.Fatal("reserved field accepted")
		}
	}
}

func TestConcurrentOutOfOrderCorrelation(t *testing.T) {
	c := helper(t, nil)
	var wg sync.WaitGroup
	failures := make(chan error, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			response, err := c.Call(ctx, NewCommand("delayed", map[string]any{"value": i, "delay": 32 - i}))
			if err != nil {
				failures <- err
				return
			}
			var data struct{ Value int }
			if err = response.DecodeData(&data); err != nil || data.Value != i {
				failures <- fmt.Errorf("got %d for %d: %v", data.Value, i, err)
			}
		}(i)
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
	if c.PendingCount() != 0 {
		t.Fatal("leaked pending requests")
	}
	if got := len(c.DrainEvents()); got != 32 {
		t.Fatalf("event stream has %d responses", got)
	}
}

func TestLargeUnicodeCRLFAndUnknownRecords(t *testing.T) {
	c := helper(t, nil)
	response := call(t, c, "large", nil)
	var data struct{ Text string }
	if err := response.DecodeData(&data); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(data.Text, "\u2028paragraph\u2029end") || len(data.Text) < 65536 {
		t.Fatal("large Unicode payload changed")
	}
	call(t, c, "malformed", nil)
	events := c.DrainEvents()
	kinds := map[string]int{}
	for _, e := range events {
		kinds[e.Type]++
		if e.Type == "future_event" && !strings.Contains(string(e.Raw), "novel") {
			t.Fatal("unknown fields lost")
		}
	}
	if kinds["protocol_error"] != 2 || kinds["future_event"] != 1 {
		t.Fatal(kinds)
	}
}

func TestCancellationCleansPendingAndIDCanBeReused(t *testing.T) {
	c := helper(t, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	command := NewCommand("hang", nil)
	command.ID = "explicit"
	if _, err := c.Call(ctx, command); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	if c.PendingCount() != 0 {
		t.Fatal("cancelled request leaked")
	}
	command.Type = GetState
	response, err := c.Call(context.Background(), command)
	if err != nil || response.ID != "explicit" {
		t.Fatal(response, err)
	}
}

func TestDuplicateIDRejectedWithoutRemovingOriginal(t *testing.T) {
	c := helper(t, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	first := make(chan error, 1)
	command := NewCommand("hang", nil)
	command.ID = "same"
	go func() { _, err := c.Call(ctx, command); first <- err }()
	deadline := time.Now().Add(time.Second)
	for c.PendingCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if _, err := c.Call(context.Background(), command); err == nil {
		t.Fatal("duplicate accepted")
	}
	if c.PendingCount() != 1 {
		t.Fatal("duplicate removed original")
	}
	cancel()
	<-first
}

func TestUIResponseIsOneWayAndPreservesID(t *testing.T) {
	c := helper(t, nil)
	if err := c.CancelDialog(context.Background(), "dialog-123"); err != nil {
		t.Fatal(err)
	}
	call(t, c, GetState, nil)
	if c.PendingCount() != 0 {
		t.Fatal("one-way dialog added pending call")
	}
	var found bool
	for _, event := range c.DrainEvents() {
		if event.Type == "ui_received" {
			found = true
			var record map[string]any
			if err := event.DecodeField("record", &record); err != nil {
				t.Fatal(err)
			}
			if record["id"] != "dialog-123" || record["cancelled"] != true || record["value"] != nil {
				t.Fatal(record)
			}
		}
	}
	if !found {
		t.Fatal("dialog response not received")
	}
	if _, err := c.Call(context.Background(), NewCommand("extension_ui_response", nil)); err == nil {
		t.Fatal("normal Call accepted one-way type")
	}
}

func TestStderrDrainedAndBounded(t *testing.T) {
	c := helper(t, func(config *Config) { config.StderrLines = 7; config.StderrLineBytes = 80 })
	call(t, c, "stderr", nil)
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	lines := c.RecentStderr()
	if len(lines) != 7 {
		t.Fatal(len(lines))
	}
	if !strings.HasPrefix(lines[6], "line-299:") {
		t.Fatal(lines[6])
	}
	for _, line := range lines {
		if len(line) > 100 || !strings.HasSuffix(line, "[truncated]") {
			t.Fatal(line)
		}
	}
}

func TestUnexpectedExitFailsPendingAndReaps(t *testing.T) {
	c := helper(t, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := c.Call(ctx, NewCommand("exit", nil)); err == nil || errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	awaitDone(t, c)
	if c.ExitCode() != 7 || c.Alive() || c.PendingCount() != 0 {
		t.Fatal("wrong exited state")
	}
	events := c.DrainEvents()
	if events[len(events)-1].Type != "process_exited" {
		t.Fatal(events)
	}
}

func TestCommandFailuresAndMismatchedResponses(t *testing.T) {
	c := helper(t, nil)
	response, err := c.Call(context.Background(), NewCommand("fail", nil))
	var rejection *CommandError
	if !errors.As(err, &rejection) || response.Success {
		t.Fatal(response, err)
	}
	if _, err = c.Call(context.Background(), NewCommand("mismatch", nil)); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatal(err)
	}
	call(t, c, GetState, nil)
}

func TestArgsEnvironmentAndCwd(t *testing.T) {
	t.Setenv("ORBIT_TEST_SECRET", "must-not-inherit")
	var directory string
	c := helper(t, func(config *Config) {
		directory = config.WorkspaceDir
		config.TrustProject = true
		config.Offline = true
		config.NoSession = true
		config.SessionDir = "session dir"
		config.Extensions = []string{"one.js", "two.js"}
		config.Args = []string{"--provider", "test provider"}
		config.EnvOverrides["ORBIT_TEST_MARKER"] = "child only"
	})
	response := call(t, c, "inspect", nil)
	var data struct {
		Args                               []string
		Cwd, Marker, Skip, Offline, Secret string
	}
	if err := response.DecodeData(&data); err != nil {
		t.Fatal(err)
	}
	expected := []string{"--mode", "rpc", "--approve", "--offline", "--no-session", "--session-dir", "session dir", "--extension", "one.js", "--extension", "two.js", "--provider", "test provider"}
	if !reflect.DeepEqual(data.Args[3:], expected) || data.Cwd != directory || data.Marker != "child only" || data.Skip != "1" || data.Offline != "1" || data.Secret != "" {
		t.Fatal(data)
	}
}

func TestEventOverflowFailsClosed(t *testing.T) {
	c := helper(t, func(config *Config) { config.MaxEventBytes = 3000 })
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := c.Call(ctx, NewCommand("flood", nil))
	if !errors.Is(err, ErrEventOverflow) {
		t.Fatal(err)
	}
	awaitDone(t, c)
	if !errors.Is(c.Err(), ErrEventOverflow) {
		t.Fatal(c.Err())
	}
}

func TestRecordLimitFailsClosed(t *testing.T) {
	c := helper(t, func(config *Config) { config.MaxRecordBytes = 1024 })
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := c.Call(ctx, NewCommand("large", nil))
	if err == nil || !strings.Contains(err.Error(), "record exceeds") {
		t.Fatal(err)
	}
	awaitDone(t, c)
}

func TestGracefulAndForcedClose(t *testing.T) {
	for _, ignoreEOF := range []bool{false, true} {
		t.Run(fmt.Sprint(ignoreEOF), func(t *testing.T) {
			c := helper(t, func(config *Config) {
				if ignoreEOF {
					config.EnvOverrides["ORBIT_RPC_IGNORE_EOF"] = "1"
				}
			})
			call(t, c, GetState, nil)
			start := time.Now()
			if err := c.Close(); err != nil {
				t.Fatal(err)
			}
			if time.Since(start) > 3*time.Second {
				t.Fatal("close hung")
			}
			if err := c.Close(); err != nil {
				t.Fatal(err)
			}
			if _, err := c.Call(context.Background(), NewCommand(GetState, nil)); !errors.Is(err, ErrClosed) {
				t.Fatal(err)
			}
		})
	}
}

func TestCancelledBlockedWriterAndClose(t *testing.T) {
	c := helper(t, func(config *Config) { config.EnvOverrides["ORBIT_RPC_NO_READ"] = "1" })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := c.Call(ctx, NewCommand("large-write", map[string]any{"data": strings.Repeat("x", 2<<20)}))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestStartFailureAndResolveOverride(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-pi")
	if _, err := Start(Config{Executable: missing}); err == nil {
		t.Fatal("missing binary launched")
	}
	t.Setenv("PI_BIN", missing)
	if ResolveBinary() != missing {
		t.Fatal("override ignored")
	}
}
