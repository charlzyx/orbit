//go:build !windows

package pirpc

import "os/exec"

func configureProcess(command *exec.Cmd) {}
func platformCommand(binary string, args []string) (string, []string, error) {
	return binary, args, nil
}
