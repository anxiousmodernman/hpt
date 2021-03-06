package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Exec maps a shelled-out command.
type Exec struct {
	Cmd    string   `toml:"cmd"`
	Args   []string `toml:"args"`
	Script string   `toml:"script"`
	// TODO passing env vars, running as a user
	Env  map[string]string `toml:"env"`
	User string            `toml:"user"`
	PWD  string            `toml:"pwd"`
}

// ApplyExec ...
func ApplyExec(conf Config, e Exec) *ApplyState {
	state := NewApplyState("exec")

	if e.Script != "" {
		out, err := execTempFile(e.Script, e.User, e.PWD)
		if err != nil {
			if out != nil {
				state.Output.Write(out)
			}
			return state.Errorf("execTempFile error: %v", err)

		}
		state.Output.Write(out)
	}

	return state
}

func execTempFile(script, as, pwd string) ([]byte, error) {

	f, err := ioutil.TempFile("", "hpt-script")
	if err != nil {
		return nil, err
	}
	defer func() { os.Remove(f.Name()) }()

	_, err = io.Copy(f, bytes.NewBufferString(script))
	if err != nil {
		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	fqp := filepath.Join(f.Name())
	err = os.Chmod(fqp, 0700)
	if err != nil {
		return nil, fmt.Errorf("chmod err: %v", err)

	}

	c := exec.Command("/bin/sh", fqp)
	if as != "" {
		u, err := user.Lookup(as)
		if err != nil {
			if _, ok := err.(user.UnknownUserError); ok {
				return nil, errors.New("user does not exist")
			}
			return nil, err
		}
		uid, err := strconv.ParseUint(u.Uid, 10, 32)
		if err != nil {
			return nil, err
		}
		gid, err := strconv.ParseUint(u.Gid, 10, 32)
		if err != nil {
			return nil, err
		}
		// we've been given a user to run as
		c.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: uint32(uid),
				Gid: uint32(gid),
			},
			Setsid: true,
		}

		if err := setUserOnFile(as, f.Name(), false); err != nil {
			return nil, err
		}
	}

	if pwd != "" {
		c.Dir = pwd
	} else {
		// we run in /tmp by default
		c.Dir = os.TempDir()
	}

	b, err := c.Output()
	if err != nil {
		// normally state.Error does this for us
		if err, ok := err.(*exec.ExitError); ok {
			return err.Stderr, err
		}
		return nil, err
	}

	return b, err
}

// ExecCommand executes the given command and arguments. Stdout and stderr are
// collected into a []byte and returned.
func ExecCommand(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).CombinedOutput()
}

// filterSpace filters spaces and newlines.
func filterSpace(input []string) []string {
	var result []string
	for _, x := range input {
		if strings.TrimSpace(x) != "" {
			result = append(result, x)
		}
	}
	return result
}

// ExecScriptTokens takes a (potentially) multi-line script block
// and returns a list of list of string. Each item in the outer list can be
// regarded as a "line" in the script. The items in the inner list are the
// command and any arguments.
func ExecScriptTokens(s string) [][]string {
	lines := strings.Split(s, "\n")
	// remove any "empty" lines
	lines = filterSpace(lines)

	var result [][]string
	for _, line := range lines {
		splitted := filterSpace(strings.Split(line, " "))
		result = append(result, splitted)
	}
	return result
}
