package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

// File ...
type File struct {
	Path string `toml:"path"`
	// Src is a resolver path
	Src    string `toml:"source"`
	Perms  int    `toml:"perms"`
	Remove bool   `toml:"remove"`
	IsDir  bool   `toml:"dir"`
	Owner  string `toml:"owner"`
	Group  string `toml:"group"`
	// TODO(cm): expected sha or md5 hash
}

// ApplyFile ...
func ApplyFile(conf Config, f File) *ApplyState {
	state := NewApplyState("file")

	// prerequisites on our File block configuration
	if !filepath.IsAbs(f.Path) {
		return state.Error(errors.New("filepath must be absolute"))
	}

	// handle remove
	if f.Remove {
		if f.IsDir {
			state = removeDir(conf, f, state)
		} else {
			state = removeFile(conf, f, state)
		}
	}

	// it's not a remove, so we create

	if f.IsDir {
		return createDir(conf, f, state)
	}

	return createFile(conf, f, state)
}

func createDir(conf Config, f File, state *ApplyState) *ApplyState {
	if !f.IsDir {
		panic("createDir only accepts directories")
	}

	perms := umaskHack(defaultPermsInt(f.Perms))
	err := os.Mkdir(f.Path, perms)
	if err != nil {

		return state.Error(err)
	}
	if f.Group != "" {
		err = setGroupOnFile(f.Group, f.Path, true)
		if err != nil {
			return state.Errorf("setGroupOnFile: %v", err)
		}
	}
	if f.Owner != "" {
		err = setUserOnFile(f.Owner, f.Path, true)
		if err != nil {
			return state.Errorf("setUserOnFile: %v", err)
		}
	}
	return state
}

func createFile(conf Config, f File, state *ApplyState) *ApplyState {
	fmt.Println("try to create file:", f)
	resolverName, path := ParseResolverPath(f.Src)
	if resolverName == "" {
		fmt.Println("resolverName", resolverName)
		fmt.Println("path", path)
		panic("local resolver unsupported")
	}
	resolver, err := BuildResolver(resolverName, conf)
	if err != nil {
		return state.Error(err)
	}
	r, err := resolver.Get(path)
	if err != nil {
		return state.Error(err)
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return state.Error(err)
	}
	perms := umaskHack(defaultPermsInt(f.Perms))
	// NOTE: We need a generalized way to open files.
	err = ioutil.WriteFile(f.Path, data, perms)
	if err != nil {
		return state.Error(err)
	}

	if f.Group != "" {
		exists, err := GroupExists(f.Group)
		if err != nil {
			return state.Errorf("group exists error: %v", err)
		}
		if !exists {
			msg := "group %s must exist to use in File block"
			return state.Error(errors.Errorf(msg, f.Group))
		}
		// TODO how to handle a recursive? We are in a create, so we probably
		// don't need to worry about chgrp -R here.
		cmd := exec.Command("chgrp", f.Group, f.Path)
		output, err := cmd.Output()
		if err != nil {
			if err, ok := err.(*exec.ExitError); ok {
				state.Output.Write(err.Stderr)
			}
			return state.Errorf("chgrp error: %v", err)
		}
		state.Output.Write(output)
	}

	if f.Owner != "" {
		exists, err := UserExists(f.Owner)
		if err != nil {
			return state.Errorf("user exists error: %v", err)
		}
		if !exists {
			msg := "owner %s must exist to use in File block"
			return state.Error(errors.Errorf(msg, f.Owner))
		}
		cmd := exec.Command("chown", f.Owner, f.Path)
		output, err := cmd.Output()
		if err != nil {
			return state.Errorf("chown error: %v", err)
		}
		state.Output.Write(output)
	}
	return state
}

func removeFile(conf Config, f File, state *ApplyState) *ApplyState {
	if !f.Remove {
		panic("never call removeFile if Remove is false")
	}
	exists, err := pathExists(f.Path)
	if err != nil {
		return state.Error(err)
	}
	if !exists {
		state.Outcome = Unchanged
		return state
	}
	err = os.Remove(f.Path)
	return state.Error(err)
}

func removeDir(conf Config, f File, state *ApplyState) *ApplyState {
	if !f.Remove {
		panic("never call removeDir if Remove is false")
	}
	exists, err := pathExists(f.Path)
	if err != nil {
		return state.Error(err)
	}
	if !exists {
		state.Outcome = Unchanged
		return state
	}
	err = os.RemoveAll(f.Path)
	return state
}

func defaultPerms(perms int) os.FileMode {
	if perms == 0 {
		return 0600
	}
	// A type conversion int -> uint32
	return os.FileMode(perms)
}

func defaultPermsInt(perms int) int {
	if perms == 0 {
		return 0600
	}
	return perms
}

// ComparePermissions perms must be passed as octal
func ComparePermissions(path string, uid, gid, perms int) (same bool, err error) {

	info, err := os.Stat(path)
	if err != nil {
		return
	}
	// We have to do this funky type switch because os.File does not magically
	// know it's filesystem implementation. Go is a cross-platform language, but
	// given a bare file there are not functions in the standard library that
	// can tell you who owns the file. Instead, you have to dig into an internal
	// Sys interface, and
	switch info.Sys().(type) {
	case *syscall.Stat_t:
		stat_t, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			err = errors.New("could not convert FileInfo.Sys")
			return
		}
		iuid, igid := stat_t.Uid, stat_t.Gid
		if int(iuid) != uid {
			fmt.Println("uids differ")
			return
		}
		if int(igid) != gid {
			fmt.Println("gids differ")
			return
		}
	default:
		err = errors.New("unsupported system")
		return
	}

	// If we got this far, uid and gid match the file. Now we check perms.
	mode := info.Mode()
	if !samePerms(perms, uint32(mode)) {
		fmt.Printf("declaring not same: %o %o", mode, perms)
		return
	}

	// compare username?

	same = true
	return
}

func samePerms(yours int, mode uint32) bool {
	if (uint32(yours) ^ mode) == 0 {
		return true
	}
	return false
}

func setGroupOnFile(group, path string, recursive bool) error {
	exists, err := GroupExists(group)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("group %s must exist to use in File block", group)
	}
	// TODO how to handle a recursive?
	cmd := exec.Command("chgrp", group, path)
	// we don't need stdout from chgrp
	_, err = cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
func setUserOnFile(user, path string, recursive bool) error {
	exists, err := UserExists(user)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("user %s must exist to use in File block", user)
	}

	cmd := exec.Command("chown", user, path)
	_, err = cmd.Output()
	if err != nil {
		fmt.Errorf("chown error: %v", err)
	}

	return nil
}

// stolen from the go language build
// https://github.com/golang/go/commit/0cb68acea2e82f6e071804d4d890271103f83c7b
func umaskHack(perms int) os.FileMode {
	tempFile := filepath.Join("/tmp", uuid.NewRandom().String())
	mode := os.FileMode(perms)
	f, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.FileMode(perms))
	if err == nil {
		fi, err := f.Stat()
		if err == nil {
			mode = fi.Mode() & 0777
		}
		name := f.Name()
		f.Close()
		os.Remove(name)
	}
	return mode
}
