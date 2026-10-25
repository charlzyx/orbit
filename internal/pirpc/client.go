package pirpc

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charlzyx/orbit/internal/runtimepath"
)

var ErrClosed = errors.New("Pi RPC process is closed")
var ErrEventOverflow = errors.New("Pi RPC event buffer exceeded its configured byte limit")

// Config describes one Pi process. A nil Environment inherits the host's env;
// an explicitly empty slice is a clean environment. Overrides are child-only.
type Config struct {
	Executable      string
	PrefixArgs      []string
	WorkspaceDir    string
	SessionDir      string
	Extensions      []string
	Args            []string
	Environment     []string
	EnvOverrides    map[string]string
	TrustProject    bool
	Offline         bool
	NoSession       bool
	MaxRecordBytes  int
	MaxEventBytes   int
	MaxEventCount   int
	StderrLines     int
	StderrLineBytes int
	ShutdownTimeout time.Duration
}

type result struct {
	response Response
	err      error
}
type outgoing struct {
	ctx  context.Context
	wire []byte
	ack  chan error
}
type pending struct {
	command string
	reply   chan result
}

// Client owns and reaps exactly one process. DrainEvents never performs I/O.
// All public methods are safe for concurrent use. Event data is immutable.
type Client struct {
	config      Config
	command     *exec.Cmd
	stdin       io.WriteCloser
	outgoing    chan outgoing
	stopWriter  chan struct{}
	done        chan struct{}
	nextID      atomic.Uint64
	closeOnce   sync.Once
	failOnce    sync.Once
	mu          sync.Mutex
	pending     map[string]pending
	events      []Event
	eventBytes  int
	stderr      []string
	terminalErr error
	closing     bool
	exitCode    int
	startedAt   time.Time
	exitedAt    time.Time
}

func Start(config Config) (*Client, error) {
	if config.Executable == "" {
		config.Executable = ResolveBinary()
	}
	if config.MaxRecordBytes <= 0 {
		config.MaxRecordBytes = 32 << 20
	}
	if config.MaxEventBytes <= 0 {
		config.MaxEventBytes = 64 << 20
	}
	if config.MaxEventCount <= 0 {
		config.MaxEventCount = 16384
	}
	if config.StderrLines <= 0 {
		config.StderrLines = 200
	}
	if config.StderrLineBytes <= 0 {
		config.StderrLineBytes = 16 << 10
	}
	if config.ShutdownTimeout <= 0 {
		config.ShutdownTimeout = 2 * time.Second
	}
	args := append([]string{}, config.PrefixArgs...)
	args = append(args, "--mode", "rpc")
	if config.TrustProject {
		args = append(args, "--approve")
	} else {
		args = append(args, "--no-approve")
	}
	if config.Offline {
		args = append(args, "--offline")
	}
	if config.NoSession {
		args = append(args, "--no-session")
	}
	if config.SessionDir != "" {
		args = append(args, "--session-dir", config.SessionDir)
	}
	for _, extension := range config.Extensions {
		args = append(args, "--extension", extension)
	}
	args = append(args, config.Args...)
	executable, args, err := platformCommand(config.Executable, args)
	if err != nil {
		return nil, err
	}
	command := exec.Command(executable, args...)
	command.Dir = config.WorkspaceDir
	command.Env = childEnvironment(config)
	configureProcess(command)
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, stdoutWriter, err := os.Pipe()
	if err != nil {
		stdin.Close()
		return nil, err
	}
	stderr, stderrWriter, err := os.Pipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		stdoutWriter.Close()
		return nil, err
	}
	command.Stdout = stdoutWriter
	command.Stderr = stderrWriter
	if err := command.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		stdoutWriter.Close()
		stderrWriter.Close()
		return nil, fmt.Errorf("start Pi: %w", err)
	}
	stdoutWriter.Close()
	stderrWriter.Close()
	c := &Client{config: config, command: command, stdin: stdin, outgoing: make(chan outgoing, 128), stopWriter: make(chan struct{}), done: make(chan struct{}), pending: map[string]pending{}, exitCode: -1, startedAt: time.Now()}
	stdoutDone, stderrDone := make(chan struct{}), make(chan struct{})
	go func() { defer close(stdoutDone); c.readStdout(stdout) }()
	go func() { defer close(stderrDone); c.readStderr(stderr) }()
	go c.writeLoop()
	go func() {
		// Owned pipe readers are not closed by Wait. Drain the final records,
		// but do not hang forever if a descendant inherited an output pipe.
		waitErr := command.Wait()
		c.mu.Lock()
		c.exitedAt = time.Now()
		c.exitCode = command.ProcessState.ExitCode()
		c.mu.Unlock()
		for _, drained := range []<-chan struct{}{stdoutDone, stderrDone} {
			select {
			case <-drained:
			case <-time.After(config.ShutdownTimeout):
				stdout.Close()
				stderr.Close()
				<-drained
			}
		}
		stdout.Close()
		stderr.Close()
		c.mu.Lock()
		c.exitCode = command.ProcessState.ExitCode()
		if c.terminalErr == nil {
			if waitErr != nil && !c.closing {
				c.terminalErr = fmt.Errorf("Pi exited: %w", waitErr)
			} else {
				c.terminalErr = ErrClosed
			}
		}
		terminalErr := c.terminalErr
		for id, request := range c.pending {
			request.reply <- result{err: terminalErr}
			delete(c.pending, id)
		}
		c.events = append(c.events, localEvent("process_exited", terminalErr.Error()))
		c.mu.Unlock()
		close(c.done)
	}()
	return c, nil
}

func childEnvironment(config Config) []string {
	env := config.Environment
	if env == nil {
		env = os.Environ()
	}
	values := make(map[string]string, len(env)+2)
	for _, entry := range env {
		if key, value, ok := strings.Cut(entry, "="); ok {
			values[key] = value
		}
	}

	values["PI_SKIP_VERSION_CHECK"] = "1"
	if config.Offline {
		values["PI_OFFLINE"] = "1"
	}
	for key, value := range config.EnvOverrides {
		values[key] = value
	}
	out := make([]string, 0, len(values))
	for key, value := range values {
		out = append(out, key+"="+value)
	}
	preferred := filepath.Dir(config.Executable)
	if preferred == "." {
		preferred = ""
	}
	return runtimepath.AugmentEnvironment(out, preferred)
}

func (c *Client) Call(ctx context.Context, command Command) (Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	if isOneWay(command.Type) {
		return Response{}, errors.New("extension UI records require SendOneWay, not Call")
	}
	if command.ID == "" {
		command.ID = fmt.Sprintf("orbit-go-%d", c.nextID.Add(1))
	}
	wire, err := json.Marshal(command)
	if err != nil {
		return Response{}, err
	}
	reply := make(chan result, 1)
	c.mu.Lock()
	if c.closing || c.terminalErr != nil {
		c.mu.Unlock()
		return Response{}, ErrClosed
	}
	if _, exists := c.pending[command.ID]; exists {
		c.mu.Unlock()
		return Response{}, fmt.Errorf("duplicate request ID %q", command.ID)
	}
	c.pending[command.ID] = pending{command: command.Type, reply: reply}
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		if p, ok := c.pending[command.ID]; ok && p.reply == reply {
			delete(c.pending, command.ID)
		}
		c.mu.Unlock()
	}()
	select {
	case c.outgoing <- outgoing{ctx: ctx, wire: wire}:
	case <-ctx.Done():
		return Response{}, ctx.Err()
	case <-c.done:
		return Response{}, c.Err()
	}
	select {
	case answer := <-reply:
		return answer.response, answer.err
	case <-ctx.Done():
		return Response{}, ctx.Err()
	case <-c.done:
		select {
		case answer := <-reply:
			return answer.response, answer.err
		default:
			return Response{}, c.Err()
		}
	}
}

func isOneWay(kind string) bool {
	return kind == "extension_ui_response" || kind == "extension_ui_input" || kind == "extension_ui_resize"
}

// SendOneWay does not create a pending response. Its return acknowledges writing
// the JSONL record, not that an extension has accepted the user's decision.
func (c *Client) SendOneWay(ctx context.Context, record any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	wire, err := json.Marshal(record)
	if err != nil {
		return err
	}
	event, err := ParseEvent(wire)
	if err != nil {
		return err
	}
	if !isOneWay(event.Type) || event.ID == "" {
		return errors.New("one-way record requires an extension UI type and request ID")
	}
	c.mu.Lock()
	closed := c.closing || c.terminalErr != nil
	c.mu.Unlock()
	if closed {
		return ErrClosed
	}
	ack := make(chan error, 1)
	select {
	case c.outgoing <- outgoing{ctx: ctx, wire: wire, ack: ack}:
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return c.Err()
	}
	select {
	case err := <-ack:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return c.Err()
	}
}

func (c *Client) SendUI(ctx context.Context, response UIResponse) error {
	response.Type = "extension_ui_response"
	return c.SendOneWay(ctx, response)
}
func (c *Client) CancelDialog(ctx context.Context, id string) error {
	return c.SendUI(ctx, UIResponse{ID: id, Cancelled: true})
}

func (c *Client) writeLoop() {
	for {
		select {
		case <-c.done:
			return
		case <-c.stopWriter:
			return
		case message := <-c.outgoing:
			if err := message.ctx.Err(); err != nil {
				if message.ack != nil {
					message.ack <- err
				}
				continue
			}
			_, err := c.stdin.Write(append(message.wire, '\n'))
			if message.ack != nil {
				message.ack <- err
			}
			if err != nil {
				c.mu.Lock()
				closing := c.closing
				c.mu.Unlock()
				if !closing {
					c.fail(fmt.Errorf("write Pi stdin: %w", err))
				}
				return
			}
		}
	}
}

func readRecord(reader *bufio.Reader, limit int) ([]byte, error) {
	var record []byte
	for {
		part, err := reader.ReadSlice('\n')
		if len(record)+len(part) > limit {
			return nil, fmt.Errorf("Pi JSONL record exceeds %d bytes", limit)
		}
		record = append(record, part...)
		if err == bufio.ErrBufferFull {
			continue
		}
		if err == io.EOF && len(record) > 0 {
			err = nil
		}
		record = bytes.TrimSuffix(record, []byte{'\n'})
		record = bytes.TrimSuffix(record, []byte{'\r'})
		return record, err
	}
}

func (c *Client) readStdout(stdout io.Reader) {
	reader := bufio.NewReader(stdout)
	for {
		record, err := readRecord(reader, c.config.MaxRecordBytes)
		if err == io.EOF {
			return
		}
		if err != nil {
			c.fail(fmt.Errorf("read Pi stdout: %w", err))
			return
		}
		if len(record) == 0 {
			continue
		}
		event, parseErr := ParseEvent(record)
		if parseErr != nil {
			raw, _ := json.Marshal(map[string]any{"type": "protocol_error", "error": parseErr.Error(), "record": string(record)})
			event, _ = ParseEvent(raw)
		}
		if !c.enqueue(event) {
			c.fail(ErrEventOverflow)
			return
		}
		if event.Type == "response" {
			var response Response
			if err := event.Decode(&response); err == nil {
				response.Raw = bytes.Clone(event.Raw)
				c.mu.Lock()
				if request, ok := c.pending[response.ID]; ok {
					delete(c.pending, response.ID)
					var callErr error
					if response.Command != request.command {
						callErr = fmt.Errorf("response command %q does not match %q", response.Command, request.command)
					} else if !response.Success {
						callErr = &CommandError{Response: response}
					}
					request.reply <- result{response: response, err: callErr}
				}
				c.mu.Unlock()
			}
		}
	}
}

func (c *Client) enqueue(event Event) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.eventBytes+len(event.Raw) > c.config.MaxEventBytes || len(c.events) >= c.config.MaxEventCount {
		return false
	}
	c.events = append(c.events, event)
	c.eventBytes += len(event.Raw)
	return true
}

func (c *Client) readStderr(stderr io.Reader) {
	reader := bufio.NewReader(stderr)
	var line []byte
	truncated := false
	for {
		part, err := reader.ReadSlice('\n')
		room := c.config.StderrLineBytes - len(line)
		if len(part) > room {
			line = append(line, part[:room]...)
			truncated = true
		} else {
			line = append(line, part...)
		}
		if err == bufio.ErrBufferFull {
			continue
		}
		if len(line) > 0 || truncated {
			value := strings.TrimRight(string(line), "\r\n")
			if truncated {
				value += " [truncated]"
			}
			c.mu.Lock()
			if len(c.stderr) == c.config.StderrLines {
				copy(c.stderr, c.stderr[1:])
				c.stderr[len(c.stderr)-1] = value
			} else {
				c.stderr = append(c.stderr, value)
			}
			c.mu.Unlock()
		}
		line = nil
		truncated = false
		if err != nil {
			return
		}
	}
}

func (c *Client) fail(err error) {
	c.failOnce.Do(func() {
		c.mu.Lock()
		c.terminalErr = err
		for id, request := range c.pending {
			request.reply <- result{err: err}
			delete(c.pending, id)
		}
		c.events = append(c.events, localEvent("transport_error", err.Error()))
		c.mu.Unlock()
		_ = c.command.Process.Kill()
	})
}

func (c *Client) DrainEvents() []Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	events := c.events
	c.events = nil
	c.eventBytes = 0
	return events
}
func (c *Client) RecentStderr() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string{}, c.stderr...)
}
func (c *Client) Done() <-chan struct{} { return c.done }
func (c *Client) PID() int              { return c.command.Process.Pid }
func (c *Client) Err() error            { c.mu.Lock(); defer c.mu.Unlock(); return c.terminalErr }
func (c *Client) ExitCode() int         { c.mu.Lock(); defer c.mu.Unlock(); return c.exitCode }
func (c *Client) PendingCount() int     { c.mu.Lock(); defer c.mu.Unlock(); return len(c.pending) }
func (c *Client) Alive() bool {
	select {
	case <-c.done:
		return false
	default:
		return true
	}
}

// Close first closes stdin for Pi's orderly disposal, then kills after the
// configured deadline. It is idempotent and always waits for process reaping.
func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closing = true
		c.mu.Unlock()
		close(c.stopWriter)
		_ = c.stdin.Close()
		select {
		case <-c.done:
		case <-time.After(c.config.ShutdownTimeout):
			_ = c.command.Process.Kill()
			<-c.done
		}
	})
	err := c.Err()
	if errors.Is(err, ErrClosed) {
		return nil
	}
	return err
}

// ResolveBinary honors Orbit's PI_BIN override, then known GUI install paths.
func ResolveBinary() string {
	if value := os.Getenv("PI_BIN"); value != "" {
		return value
	}
	if value := runtimepath.Resolve("pi"); value != "" {
		return value
	}
	return "pi"
}
