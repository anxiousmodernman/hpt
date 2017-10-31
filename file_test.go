package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateFileWithPermissions(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("unfortunately, this test must run as root")
	}

	// We expect group testgroup to exist
	// We expect user test to exist
	if exists, _ := UserExists(testUser); !exists {
		t.Fatalf("test prerequisite user %v must exist", testUser)
	}
	if exists, _ := GroupExists(testGroup); !exists {
		t.Fatalf("test prerequisite group %v must exist", testGroup)
	}

	cases := []struct {
		name  string
		perms int
		user  string
		group string
	}{
		{
			name:  "test1",
			perms: 0644,
			user:  testUser,
			group: testGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			tempDir := filepath.Join("/tmp", tc.name+"dir")

			//defer func() { os.RemoveAll(tempDir) }()

			// We need a Config for our ApplyFile function, but because we use a
			// test resolver, the contents of the config can be blank.
			var conf Config

			// make the dir first
			var dir File
			dir.Group = tc.group
			dir.Owner = tc.user
			dir.Path = tempDir
			dir.Src = "" // no op for directories
			dir.Perms = 0750
			dir.IsDir = true

			state := ApplyFile(conf, dir)
			if state.Err != nil {
				t.Logf("output: %s", readAllStr(state.Output))
				t.Errorf("ApplyFile error: %v", state.Err)
			}

			var f File
			f.Group = tc.group
			f.Owner = tc.user
			// we don't provide a path in our test cases because we generate a
			// temp path and assign it here
			f.Path = filepath.Join(tempDir, tc.name)
			// We fake the Src here and signal BuildResolver to use an in-memory
			// TestResolver that will always give us some bytes to write to
			// a file. Here do not test the contents or checksum, only that
			// a file is written with the correct permissions.
			f.Src = "test://notimportant"
			f.Perms = tc.perms

			state = ApplyFile(conf, f)
			if state.Err != nil {
				t.Logf("output: %s", readAllStr(state.Output))
				t.Errorf("ApplyFile error: %v", state.Err)
			}
		})
	}
}

func readAllStr(r io.Reader) string {
	if r == nil {
		return "nil io.Reader"
	}
	everything, _ := ioutil.ReadAll(r)
	return string(everything)
}
