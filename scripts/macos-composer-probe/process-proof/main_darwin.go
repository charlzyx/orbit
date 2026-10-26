//go:build darwin && arm64

package main

import (
	"encoding/json"
	proof "github.com/charlzyx/orbit/scripts/macos-composer-probe/menuproof"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) == 5 && os.Args[1] == "window-proof" {
		pid, e := strconv.Atoi(os.Args[2])
		if e != nil {
			os.Exit(2)
		}
		start, e := strconv.ParseUint(os.Args[3], 10, 64)
		if e != nil {
			os.Exit(2)
		}
		number, e := strconv.ParseUint(os.Args[4], 10, 32)
		if e != nil {
			os.Exit(2)
		}
		p, e := proof.InspectWindow(pid, start, uint32(number))
		if e != nil {
			os.Exit(3)
		}
		if json.NewEncoder(os.Stdout).Encode(p) != nil {
			os.Exit(4)
		}
		return
	}
	if len(os.Args) != 3 || os.Args[1] != "process-proof" {
		os.Exit(2)
	}
	pid, e := strconv.Atoi(os.Args[2])
	if e != nil {
		os.Exit(2)
	}
	p, e := proof.InspectProcess(pid)
	if e != nil {
		os.Exit(3)
	}
	if json.NewEncoder(os.Stdout).Encode(p) != nil {
		os.Exit(4)
	}
}
