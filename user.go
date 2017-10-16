package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
)

type UserState struct {
	Changed bool
	Err     error
	Message string
	Output  io.Reader
}

// ApplyUser ...
func ApplyUser(u User) (*UserState, error) {

	output := bytes.NewBuffer([]byte(""))
	_ = output
	var us UserState
	_, err := user.Lookup(u.Name)
	if err != nil {
		if _, ok := err.(user.UnknownUserError); ok {
			log.Printf("user %s does not exist", u.Name)
			goto NOEXIST
		}
		return nil, err
	}

NOEXIST:
	// make user
	cmd := exec.Command("adduser", u.Name)
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	// make home
	us.Changed = true

	exists, err := pathExists(u.Home)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := os.Mkdir(u.Home, 0755); err != nil {
			return nil, fmt.Errorf("mkdir fail: %v", err)
		}
	}

	_, uid, gid, err := lookupUser(u.Name)
	if err != nil {
		// could be UnknownUserError, but at this point
		// that would indicate an unrecoverable error
		return nil, err
	}

	err = os.Chown(u.Home, int(uid), int(gid))
	if err != nil {
		return nil, fmt.Errorf("chown home: %v", err)
	}
	// make ssh dir 700
	sshDir := filepath.Join(u.Home, ".ssh")
	exists, err = pathExists(sshDir)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := os.Mkdir(sshDir, 0755); err != nil {
			return nil, fmt.Errorf("mkdir fail: %v", err)
		}
	}

	// parse a path
	// resolver, path

	// get ssh public keys
	// pub 644
	// priv 600

	return &us, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func lookupUser(name string) (*user.User, int64, int64, error) {
	u, err := user.Lookup(name)
	if err != nil {
		return nil, 0, 0, err
	}
	uid, err := strconv.ParseInt(u.Uid, 10, 32)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("parse uid: %v", err)
	}
	gid, err := strconv.ParseInt(u.Gid, 10, 32)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("parse uid: %v", err)
	}

	return u, uid, gid, nil
}
