package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

// Group ...
type Group struct {
	Name             string `toml:"name"`
	GID              string `toml:"gid"`
	PasswordlessSudo bool   `toml:"passwordless_sudo"`
	Absent           bool   `toml:"absent"`
}

// ApplyGroup ...
func ApplyGroup(g Group, conf Config) *ApplyState {
	// NOTE: this function sucks and needs refactoring.
	state := NewApplyState("group")
	if g.Absent {
		return removeGroup(g, conf, state)
	}

	_, err := user.LookupGroup(g.Name)
	if err != nil {
		if _, ok := err.(user.UnknownGroupError); ok {
			// not an error
			log.Printf("group %s does not exist", g.Name)
			var args []string
			// if gid provided, pass it on command line
			if strings.TrimSpace(g.GID) != "" {
				args = append(args, "--gid", g.GID)
			}
			args = append(args, g.Name)

			cmd := exec.Command("groupadd", args...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				state.Output = bytes.NewBuffer(out)
				return state.Error(fmt.Errorf("groupadd: %v", err))
			}
			state.Output = bytes.NewBuffer(out)
			state.Outcome = Changed
			fmt.Println("created group", g.Name)
		} else {
			return state.Error(err)
		}
	}

	if g.PasswordlessSudo {
		err := passwordlessSudo(g.Name)
		if err != nil {
			return state.Error(err)
		}
	}

	return state
}

func removeGroup(g Group, conf Config, state *ApplyState) *ApplyState {
	cmd := fmt.Sprintf("groupdel %s", g.Name)
	c := exec.Command(cmd)
	out, err := c.Output()
	if err != nil {
		state.Errorf("error deleting group: %v", err)
	}
	state.Output.Write(out)
	return state
}

func passwordlessSudo(name string) error {
	f, err := os.OpenFile(
		"/etc/sudoers",
		os.O_APPEND|os.O_WRONLY,
		0440)
	if err != nil {
		return fmt.Errorf("open /etc/sudoers: %v", err)
	}
	defer f.Close()
	line := "%" + fmt.Sprintf("%s        ALL=(ALL)       NOPASSWD: ALL \n", name)
	f.Write([]byte(line))
	return nil
}
