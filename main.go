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
			Name:  "manage",
			Flags: []cli.Flag{confFlag},
			Usage: "bring a box under management",
			Action: func(ctx *cli.Context) error {
				return nil
			},
		},
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
			return errors.New("you must provide a state file")
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
	yellow  = color.New(color.FgYellow)
)

// run takes a sequence of paths to config files.
func run(paths ...string) error {

	// TODO support multiple configs
	path := paths[0]

	printStates := func(stage string, as []*ApplyState) {
		for _, s := range as {
			if s.Err != nil {
				boldRed.Printf("%s apply error: %v\n", stage, s.Err)
			}
			output, err := ioutil.ReadAll(s.Output)
			if err != nil {
				boldRed.Print("could not read output\n")
				continue
			}
			white.Println(string(output))
		}
	}
	conf, err := NewConfig(path)
	if err != nil {
		return err
	}
	ep, err := NewExecutionPlan(conf)
	if err != nil {
		return err
	}
	var as []*ApplyState
	for {
		fn := ep.Next()
		if fn == nil {
			break
		}
		state := fn()
		as = append(as, state)
	}
	printStates("everything:", as)

	return nil
}

// plan takes a sequence of paths to config files.
func plan(paths ...string) error {
	return nil
}
