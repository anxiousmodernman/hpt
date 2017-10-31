package main

import (
	"fmt"
	"testing"
)

func TestExecScriptConfig(t *testing.T) {

	cases := []struct {
		config   string
		i, lines int
	}{
		{
			config: "script.toml",
			i:      0,
			lines:  3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.config, func(t *testing.T) {
			conf, _ := NewConfig(fmt.Sprintf("testfixtures/%s", tc.config))
			var count int
			script := conf.Execs[tc.i].Script
			lines := ExecScriptTokens(script)
			t.Log(script)
			fmt.Println(lines[0][0], lines[0][1:])

			if len(lines) != tc.lines {
				t.Errorf("expected %v lines got %v", tc.lines, count)
			}
		})
	}
}
