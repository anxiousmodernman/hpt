package main

import "os/exec"

// ExecCommand executes the given command and arguments. Stdout and stderr are
// collected into a []byte and returned.
func ExecCommand(cmd string, args ...string) ([]byte, error) {
	c := exec.Command(cmd, args...)
	return c.CombinedOutput()
}
