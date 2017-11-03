package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

// State represents the ultimate state of an apply call.
type State string

// String implements Stringer.
func (s State) String() string { return string(s) }

// possible states
const (
	Unchanged State = "unchanged"
	Changed   State = "changed"
)

// ApplyState is the return type of all our apply functions. A
type ApplyState struct {
	// Err is set if an error occurs in an apply function
	Err error
	// Name will take the name of the configuration block. E.g. "user" for
	// a [[user]] block
	Name string
	// Outcome is set to changed if any stateful function is called
	Outcome State
	// Output is a buffer where we can collect the output of subprocesses
	Output *bytes.Buffer
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

// RenderShell lets us print the outcome of each apply.
func (a *ApplyState) RenderShell() []byte {

	/*
		+------------------------
		 NAME
		 OUTCOME
		 ATTR: FOO
		 ATTR: BAZ
		 ERROR: <the err.Error() output>
		 OUTPUT: <buffer>
	*/

	output := bytes.NewBuffer([]byte("\n+------------------------\n"))

	write := func(key string, value interface{}) {
		if value == nil {
			return
		}
		switch value.(type) {
		case State:
			output.WriteString(fmt.Sprintf(" %s: %v\n", key, value))
		case string:
			output.WriteString(fmt.Sprintf(" %s: %s\n", key, value))
		case error:
			output.WriteString(fmt.Sprintf(" %s: %v\n", key, value))
		case *bytes.Buffer:
			output.WriteString(fmt.Sprintf(" %s: \n", key))
			buf := value.(*bytes.Buffer)
			output.Write(buf.Bytes())
		}
	}

	write("NAME", a.Name)
	write("OUTCOME", a.Outcome)
	write("ERROR", a.Err)
	write("OUTPUT", a.Output)
	return output.Bytes()
}
