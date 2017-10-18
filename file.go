package main

import (
	"io/ioutil"
	"os"
	"strings"
)

func replaceInFile(path, find, replace string) error {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	input, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	// TODO: confirm that this does not mess up /etc/sudoers
	mode := fi.Mode().Perm()

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, find) {
			lines[i] = replace
		}
	}
	output := strings.Join(lines, "\n")
	return ioutil.WriteFile(path, []byte(output), mode)
}
