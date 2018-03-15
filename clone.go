package main

import (
	"fmt"

	git "gopkg.in/src-d/go-git.v4"
)

// Clone maps the `[[clone]]` block
type Clone struct {
	URL  string `toml:"url"`
	Dest string `toml:"dest"`
	// TODO User, Key, Password don't do anything yet
	User     string `toml:"user"`
	Key      string `toml:"key"`
	Password string `toml:"password"`
}

// ApplyClone clones a repository to disk.
func ApplyClone(conf Config, repo Clone) *ApplyState {
	state := NewApplyState("clone")

	r, err := git.PlainClone(repo.Dest, false, &git.CloneOptions{
		URL:               repo.URL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		SingleBranch:      true,
	})
	// TODO: explore new api
	_, _ = r.Head()

	if err != nil {
		return state.Errorf("error cloning repo: %v", err)
	}

	exists, err := pathExists(repo.Dest)
	if err != nil {
		return state.Error(err)
	}
	if !exists {
		return state.Error(fmt.Errorf(
			"expected path to exist after clone: %s", repo.Dest))
	}

	return state
}
