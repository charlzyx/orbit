//go:build windows

package pirpc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func configureProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
func platformCommand(binary string, args []string) (string, []string, error) {
	ext := strings.ToLower(filepath.Ext(binary))
	if ext != ".cmd" && ext != ".bat" && ext != ".ps1" {
		return binary, args, nil
	}
	// Avoid shell quoting/injection: run npm's adjacent Pi JS entry with Node.
	entry := filepath.Join(filepath.Dir(binary), "node_modules", "@earendil-works", "pi-coding-agent", "dist", "bundle", "cli.js")
	if _, err := os.Stat(entry); err != nil {
		return "", nil, fmt.Errorf("Pi shell shim has no adjacent package; configure PI_BIN to a native executable: %w", err)
	}
	node, err := exec.LookPath("node.exe")
	if err != nil {
		return "", nil, err
	}
	return node, append([]string{entry}, args...), nil
}
