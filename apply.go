package main

import (
	"bytes"
)

type State int

const (
	Unchanged State = 0
	Changed   State = 1
)

type ApplyState struct {
	Output  *bytes.Buffer
	Err     error
	Outcome State
}
