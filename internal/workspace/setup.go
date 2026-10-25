package workspace

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type SetupOutcome struct {
	Script    string `json:"script"`
	Success   bool   `json:"success"`
	ExitCode  *int   `json:"exitCode,omitempty"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	Cancelled bool   `json:"cancelled"`
}

func (o SetupOutcome) Output() string {
	if strings.TrimSpace(o.Stdout) == "" {
		return o.Stderr
	}
	if strings.TrimSpace(o.Stderr) == "" {
		return o.Stdout
	}
	return strings.TrimRight(o.Stdout, "\n") + "\n\n" + strings.TrimRight(o.Stderr, "\n")
}

// RunSetupScript never deletes a worktree when setup fails. Confirmation is
// explicit because repository scripts can run arbitrary commands.
func RunSetupScript(ctx context.Context, script, root, worktree string, confirmed bool) SetupOutcome {
	o := SetupOutcome{Script: script}
	if !confirmed {
		o.Stderr = ErrConfirmationRequired.Error()
		return o
	}
	if err := ctx.Err(); err != nil {
		o.Stderr = err.Error()
		o.Cancelled = true
		return o
	}
	fi, err := os.Stat(script)
	if err != nil {
		o.Stderr = err.Error()
		return o
	}
	if !fi.Mode().IsRegular() {
		o.Stderr = "setup script is not a regular file"
		return o
	}
	shell, args := "/bin/sh", []string{script}
	if runtime.GOOS == "windows" {
		shell = "cmd"
		args = []string{"/C", script}
	}
	cmd := exec.CommandContext(ctx, shell, args...)
	configureSetupCancellation(cmd)
	cmd.Dir = worktree
	cmd.Env = append(os.Environ(), "ORBIT_ROOT_PATH="+root, "ORBIT_WORKTREE_PATH="+worktree, "GIT_TERMINAL_PROMPT=0")
	cmd.WaitDelay = 2 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	o.Stdout = stdout.String()
	o.Stderr = stderr.String()
	if err == nil {
		code := 0
		o.ExitCode = &code
		o.Success = true
		return o
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		code := ee.ExitCode()
		o.ExitCode = &code
	}
	if ctx.Err() != nil {
		o.Cancelled = true
		o.Stderr += ctx.Err().Error()
	} else if o.Stderr == "" {
		o.Stderr = err.Error()
	}
	return o
}
