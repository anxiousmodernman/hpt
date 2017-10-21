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

func ApplyClones(conf Config) []*ApplyState {

	var results []*ApplyState
	for _, c := range conf.Clones {
		state := ApplyClone(conf, c)
		results = append(results, state)
	}
	return results
}

// ApplyClone clones a repository to disk.
func ApplyClone(conf Config, repo Clone) *ApplyState {
	var state ApplyState
	state.Output = bytes.NewBuffer([]byte("clone: " + repo.URL))
	// c := githttp.DefaultClient
	//	am, err := ssh.NewPublicKeysFromFile("anxiousmodernman", "/")
	// opts := git.CloneOptions{
	// 	URL:               repo.URL,
	// 	SingleBranch:      true,
	// 	RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	// 	Auth:              http.NewBasicAuth(repo.User, repo.Password),
	// }
	//
	r, err := git.PlainClone(repo.Dest, false, &git.CloneOptions{
		URL:               repo.URL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	_ = r
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
