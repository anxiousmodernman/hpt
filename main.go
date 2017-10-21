package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// Version can be set by build flags.
var Version = "0.0.1"

var (
	confPath = flag.String("conf", "", "path to config file")
)

type Path string

// automation for small infrastructures
func main() {

	app := cli.NewApp()
	app.Version = Version

	confFlag := cli.StringFlag{
		Name:  "conf",
		Usage: "path to config",
	}

	// a command
	app.Commands = []cli.Command{
		cli.Command{
			Name:  "plan",
			Flags: []cli.Flag{confFlag},
			Action: func(ctx *cli.Context) error {
				return nil
			},
		},
	}

	// default action
	app.Action = func(ctx *cli.Context) error {
		args := ctx.Args()

		if !args.Present() {
			return errors.New("you must provide a config file")
		}

		var paths = []string{args.First()}
		paths = append(paths, args.Tail()...)

		return run(paths...)
	}

	app.Run(os.Args)

}

var (
	boldRed = color.New(color.FgRed).Add(color.Bold)
	white   = color.New(color.FgWhite)
)

// run takes a sequence of paths to config files.
func run(paths ...string) error {

	// TODO support merging multiple configs. For now, take the first.
	path := paths[0]

	conf, err := NewConfig(path)
	if err != nil {
		return err
	}

	printStates := func(stage string, as []*ApplyState) {
		for _, s := range as {
			if s.Err != nil {
				boldRed.Printf("%s apply error: %v\n", stage, s.Err)
				// We proceed, because we expect to read from our Output buffer.
			}
			output, err := ioutil.ReadAll(s.Output)
			if err != nil {
				boldRed.Print("could not read output\n")
				continue
			}
			white.Println(string(output))
		}
	}
	// ApplyGroups
	states := ApplyGroups(conf)
	printStates("groups", states)

	// ApplyUsers
	states = ApplyUsers(conf)
	printStates("users", states)

	// ApplyFiles
	// ApplyPackages
	states = ApplyPackages(conf)
	printStates("packages", states)
	// ApplyGitClone
	states = ApplyClones(conf)
	printStates("clones", states)
	// ApplyServices
	states = ApplyServices(conf)
	printStates("services", states)
	// ApplyExec

	return nil
}

// plan takes a sequence of paths to config files.
func plan(paths ...string) error {
	return nil
}
