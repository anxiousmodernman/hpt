package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

// State represents the ultimate state of an apply call.
type State int

// possible states
const (
	Unchanged State = 0
	Changed   State = 1
)

// ApplyState is the return type of all our applies.
type ApplyState struct {
	Output  *bytes.Buffer
	Err     error
	Outcome State
}

// Error lets us concisely set an error on our ApplyState as a tail call in
// one of our apply functions.
func (a *ApplyState) Error(err error) *ApplyState {
	// we check for errors from package exec becuase we need to capture the
	// output written to stderr from failed commands
	if err, ok := err.(*exec.ExitError); ok {
		if a.Output != nil {
			a.Output.Write(err.Stderr)
		}
	}
	a.Err = err
	return a
}

// Errorf is like Error but takes a format string.
func (a *ApplyState) Errorf(format string, err error) *ApplyState {
	a.Err = fmt.Errorf(format, err)
	return a
}
