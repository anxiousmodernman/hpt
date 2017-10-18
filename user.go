package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
)

// UserState ...
type UserState struct {
	Changed bool
	Err     error
	Message string
	Output  io.Reader
}

// ApplyUsers ...
func ApplyUsers(conf Config) []*ApplyState {
	var result []*ApplyState
	for _, u := range conf.Users {
		state, _ := ApplyUser(u, conf)
		result = append(result, state)
	}
	return result, nil
}

// ApplyUser ...
func ApplyUser(u User, conf Config) (*ApplyState, error) {

	var state ApplyState
	_, err := user.Lookup(u.Name)
	if err != nil {
		if _, ok := err.(user.UnknownUserError); ok {
			// not an error
			log.Printf("user %s does not exist", u.Name)
		} else {
			return nil, err
		}
	}

	// make user
	cmd := exec.Command("adduser", u.Name)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// make home
	state.State = Changed

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
	resolverName, path := ParseResolverPath(u.SSHPublicKey)
	if resolverName == "" {
		// path is local, but we don't care about that...yet
	}
	// look up resolver
	resolver, err := BuildResolver(resolverName, conf)
	if err != nil {
		return nil, err
	}
	r, err := resolver.Get(path)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, fmt.Errorf("path %s yielded no data", u.SSHPublicKey)
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read error: %v", err)
	}
	// write the public key
	err = ioutil.WriteFile(filepath.Join(u.Home, ".ssh", "id_rsa.pub"),
		data, 0644)
	if err != nil {
		return nil, err
	}

	// Append public key to authorized keys. Open with O_APPEND and all data
	// goes to the end.
	f, err := os.OpenFile(
		filepath.Join(u.Home, ".ssh", "authorized_keys"),
		os.O_APPEND|os.O_WRONLY|os.O_CREATE,
		0644)
	if err != nil {
		return nil, fmt.Errorf("open authorized_keys: %v", err)
	}
	defer f.Close()
	// append to
	f.Write(data)

	// TODO: private key
	// pub 644
	// authorized_keys 644
	// priv 600
	type chown struct {
		path  string
		perms int
	}

	toChown := []chown{
		{filepath.Join(u.Home), 0755},
		{filepath.Join(u.Home, ".ssh"), 0755},
		{filepath.Join(u.Home, ".ssh", "id_rsa.pub"), 0644},
		{filepath.Join(u.Home, ".ssh", "authorized_keys"), 0644},
	}

	for _, tc := range toChown {
		if err := os.Chown(u.Home, int(uid), int(gid)); err != nil {
			return nil, fmt.Errorf("chown %s: %v", tc.path, err)
		}
		if err := os.Chmod(tc.path, os.FileMode(tc.perms)); err != nil {
			return nil, fmt.Errorf("chmod %s: %v", tc.path, err)
		}
	}

	fmt.Println("created user", u.Name)
	return &state, nil
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
