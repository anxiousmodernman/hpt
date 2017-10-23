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
}

// ApplyGroup ...
func ApplyGroup(g Group, conf Config) *ApplyState {

	var state ApplyState
	state.Output = bytes.NewBuffer([]byte(""))
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
		fmt.Println("applied passwordless sudo")
	}

	return &state
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
	fmt.Println("appending:", line)
	f.Write([]byte(line))

	return nil
}
