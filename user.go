package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
)

// User ...
type User struct {
	Name          string   `toml:"name"`
	Home          string   `toml:"home"`
	Groups        []string `toml:"groups"`
	SSHPublicKey  string   `toml:"ssh_public_key"`
	SSHPrivateKey string   `toml:"ssh_private_key"`
	Shell         string   `toml:"shell"`
	Sudoers       bool     `toml:"sudoers"`
	Absent        bool     `toml:"absent"`
}

// ApplyUser ...
func ApplyUser(u User, conf Config) *ApplyState {
	var state *ApplyState
	state.Output = bytes.NewBuffer([]byte(""))

	exists, err := userExists(u.Name)
	if err != nil {
		return state.Error(err)
	}
	if !exists && u.Absent {
		state.Outcome = Unchanged
		state.Output.Write([]byte("nothing to do"))
		return state
	}

	state = createUser(conf, u, state)

	return state
}

func createUser(conf Config, u User, state *ApplyState) *ApplyState {
	output := bytes.NewBuffer([]byte(""))
	state.Output = output
	exists, err := userExists(u.Name)
	if err != nil {
		return state.Error(err)
	}

	if !exists {
		if u.Absent {
			state.Outcome = Unchanged
			return state
		}
		// make user
		cmd := exec.Command("adduser", u.Name)
		out, err := cmd.CombinedOutput()
		if err != nil {
			state.Output = bytes.NewBuffer(out)
			return state.Error(err)
		}
		state.Output = bytes.NewBuffer(out)
		state.Outcome = Changed
	}

	// make home

	pexists, err := pathExists(u.Home)
	if err != nil {
		return state.Error(err)
	}
	if !pexists {
		if err := os.Mkdir(u.Home, 0755); err != nil {
			state.Err = fmt.Errorf("mkdir fail: %v", err)
			return state.Error(err)
		}
	}

	_, uid, gid, err := lookupUser(u.Name)
	if err != nil {
		return state.Error(err)
	}

	err = os.Chown(u.Home, int(uid), int(gid))
	if err != nil {
		return state.Error(fmt.Errorf("chown home: %v", err))
	}
	sshDir := filepath.Join(u.Home, ".ssh")
	exists, err = pathExists(sshDir)
	if err != nil {
		return state.Error(err)
	}
	if !exists {
		if err := os.Mkdir(sshDir, 0755); err != nil {
			return state.Error(fmt.Errorf("mkdir fail: %v", err))
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
		return state.Error(err)
	}
	r, err := resolver.Get(path)
	if err != nil {
		return state.Error(err)
	}
	if r == nil {
		return state.Error(fmt.Errorf("path %s yielded no data", u.SSHPublicKey))
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return state.Error(fmt.Errorf("read error: %v", err))
	}
	// write the public key
	err = ioutil.WriteFile(filepath.Join(u.Home, ".ssh", "id_rsa.pub"),
		data, 0644)
	if err != nil {
		return state.Error(err)
	}

	// Append public key to authorized keys. Open with O_APPEND and all data
	// goes to the end.
	f, err := os.OpenFile(
		filepath.Join(u.Home, ".ssh", "authorized_keys"),
		os.O_APPEND|os.O_WRONLY|os.O_CREATE,
		0644)
	if err != nil {
		return state.Error(fmt.Errorf("open authorized_keys: %v", err))
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
			return state.Error(fmt.Errorf("chown %s: %v", tc.path, err))
		}
		if err := os.Chmod(tc.path, os.FileMode(tc.perms)); err != nil {
			return state.Error(fmt.Errorf("chmod %s: %v", tc.path, err))
		}
	}

	fmt.Println("created user", u.Name)

	// add users to groups
	// groups must exist

	if len(u.Groups) > 0 {
		for _, grp := range u.Groups {
			cmd := exec.Command("usermod", "-aG", grp, u.Name)
			out, err := cmd.CombinedOutput()
			if err != nil {
				return state.Error(fmt.Errorf("usermod: %v", err))
			}
			output.Write(output)
			fmt.Println("added group", grp)
		}
	}

	return state
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

func userExists(name string) (bool, error) {
	_, err := user.Lookup(name)
	if err != nil {
		if _, ok := err.(user.UnknownUserError); ok {
			// the user does not exist, swallow the error
			return false, nil
		}
		return false, err
	}
	return true, nil
}
