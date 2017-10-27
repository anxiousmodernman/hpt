package main

import (
	"bytes"
	"fmt"

	git "gopkg.in/src-d/go-git.v4"
)

type Clone struct {
	URL      string `toml:"url"`
	Dest     string `toml:"dest"`
	User     string `toml:"user"`
	Key      string `toml:"key"`
	Password string `toml:"password"`
}

// ApplyClone clones a repository to disk.
func ApplyClone(conf Config, repo Clone) *ApplyState {
	var state ApplyState
	fmt.Println("cloning", repo)
	state.Output = bytes.NewBuffer([]byte("clone: " + repo.URL))

	r, err := git.NewFilesystemRepository(repo.Dest)
	if err != nil {
		return state.Errorf("error creating repo: %v", err)
	}
	err = r.Clone(&git.CloneOptions{
		URL:          repo.URL,
		SingleBranch: true,
	})
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

	return &state
}
