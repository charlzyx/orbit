//go:build windows

package native

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/runtimepath"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func runtimeProgram(binary string, args []string) (string, []string, error) {
	return runtimeProgramIn(binary, args, os.Environ())
}
func runtimeProgramIn(binary string, args []string, environment []string) (string, []string, error) {
	ext := strings.ToLower(filepath.Ext(binary))
	if ext != ".cmd" && ext != ".bat" && ext != ".ps1" {
		return binary, args, nil
	}
	root := filepath.Join(filepath.Dir(binary), "node_modules", "@earendil-works", "pi-coding-agent")
	if !isPiPackage(root) {
		return "", nil, fmt.Errorf("cannot execute an unrecognized shell shim without invoking a command shell: %s", binary)
	}
	for _, relative := range []string{"dist/bundle/cli.js", "dist/cli.js"} {
		entry := filepath.Join(root, filepath.FromSlash(relative))
		if info, err := os.Stat(entry); err == nil && !info.IsDir() {
			node := runtimepath.ResolveIn("node", environment, runtime.GOOS)
			if node == "" || strings.ToLower(filepath.Ext(node)) != ".exe" {
				return "", nil, fmt.Errorf("pi requires a native node.exe executable")
			}
			return node, append([]string{entry}, args...), nil
		}
	}
	return "", nil, fmt.Errorf("Pi shim has no adjacent JS entrypoint: %s", binary)
}
func detachRuntimeProcess(cmd *exec.Cmd) { hideConsole(cmd) }

// CommandContext terminates the owned Windows process. WaitDelay in the
// common runner also bounds inherited pipes; no unrelated process is killed.
func configureRuntimeCancellation(cmd *exec.Cmd) {}
