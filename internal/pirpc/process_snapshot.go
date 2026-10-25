package pirpc

import "time"

// ProcessSnapshot contains only observed process metadata, never arguments or
// environment. Executable is os/exec's actual command path; Launcher is the
// selected Pi entry point (these differ for Windows npm/Node launchers).
type ProcessSnapshot struct {
	PID                  int
	Executable, Launcher string
	StartedAt, ExitedAt  time.Time
	Alive                bool
	ExitCode             *int
}

func (c *Client) ProcessSnapshot() ProcessSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	value := ProcessSnapshot{PID: c.command.Process.Pid, Executable: c.command.Path, Launcher: c.config.Executable, StartedAt: c.startedAt, ExitedAt: c.exitedAt, Alive: c.exitedAt.IsZero()}
	if !c.exitedAt.IsZero() {
		code := c.exitCode
		value.ExitCode = &code
	}
	return value
}
