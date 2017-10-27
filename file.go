package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
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
	state := &ApplyState{}
	state.Output = bytes.NewBuffer([]byte("file apply:"))

	// prerequisites
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

	if f.IsDir {
		return createDir(conf, f, state)
	}

	// it's not a remove, so we create
	return createFile(conf, f, state)
}

func createDir(conf Config, f File, state *ApplyState) *ApplyState {
	if !f.IsDir {
		panic("createDir only accepts directories")
	}
	err := os.Mkdir(f.Path, defaultPerms(f.Perms))
	return state.Error(err)
}

func createFile(conf Config, f File, state *ApplyState) *ApplyState {
	resolverName, path := ParseResolverPath(f.Src)
	if resolverName == "" {
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
	err = ioutil.WriteFile(f.Path, data, defaultPerms(f.Perms))
	if err != nil {
		return state.Error(err)
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
