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
	Name    string
	Output  *bytes.Buffer
	Err     error
	Outcome State
}

// NewApplyState is a convenience constructor for an ApplyState.
func NewApplyState(name string) *ApplyState {
	a := &ApplyState{
		Name:   name,
		Output: bytes.NewBuffer([]byte("")),
	}
	return a
}

// Error lets us concisely set an error on our ApplyState as a tail call in
// one of our apply functions.
func (a *ApplyState) Error(err error) *ApplyState {
	// we check for errors from package exec because we need to capture the
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
