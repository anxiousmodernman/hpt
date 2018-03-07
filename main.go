package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// Version can be set by build flags.
var Version = "0.0.1"

var (
	confPath = flag.String("conf", "", "path to config file")
)

func main() {

	app := cli.NewApp()
	app.Version = Version

	confFlag := cli.StringFlag{
		Name:  "conf",
		Usage: "path to config",
	}

	sshUser := cli.StringFlag{
		Name:  "user",
		Usage: "ssh user to connect as",
	}

	sshPrivKeyPath := cli.StringFlag{
		Name:  "sshIdent",
		Usage: "path to private ssh key",
	}

	doAccessKey := cli.StringFlag{
		Name:   "doAccessKey",
		Usage:  "DigitalOcean access key; useful for accessing buckets",
		EnvVar: "DO_ACCESS_KEY",
	}
	doSecretAccessKey := cli.StringFlag{
		Name:   "doAccessKey",
		Usage:  "DigitalOcean SECRET access key; useful for accessing buckets",
		EnvVar: "DO_SECRET_ACCESS_KEY",
	}
	keystorePath := cli.StringFlag{
		Name:  "keystore",
		Usage: "Path to keystore for a target's hpt instance",
		Value: "/etc/hpt/keystore.db",
	}
	serverPort := cli.StringFlag{
		Name:  "port",
		Usage: "Port to listen on",
		Value: "6632", // GGEZ
	}

	app.Commands = []cli.Command{
		cli.Command{
			// TODO deprecate in favor of grpc
			Name: "apply-ssh",
			Flags: []cli.Flag{confFlag, sshUser, sshPrivKeyPath, doAccessKey,
				doSecretAccessKey},
			Usage: "run an hpt config over ssh",
			Action: func(ctx *cli.Context) error {
				if !ctx.Args().Present() {
					fmt.Println("you must provide an hpt config")
					os.Exit(1)
				}
				user, key := ctx.String("user"), ctx.String("sshIdent")
				err := ApplySSH(ctx.Args().First(), ctx.Args().Get(1), user, key)
				return err
			},
		},
		cli.Command{
			// TODO deprecate
			Name: "manage",
			Flags: []cli.Flag{confFlag, sshUser, sshPrivKeyPath, doAccessKey,
				doSecretAccessKey},
			Usage: "bring a box under management",
			Action: func(ctx *cli.Context) error {
				if !ctx.Args().Present() {
					fmt.Println("you must provide an IP to manage")
					os.Exit(1)
				}
				user, key := ctx.String("user"), ctx.String("sshIdent")
				err := Manage(ctx.Args().First(), user, key)
				return err
			},
		},
		cli.Command{
			Name:  "plan",
			Flags: []cli.Flag{confFlag},
			Action: func(ctx *cli.Context) error {
				// TODO
				return nil
			},
		},
		cli.Command{
			Name: "serve",
			// serve may be always up, or it may be a socket-activated service
			Flags: []cli.Flag{keystorePath, serverPort},
			Usage: "start an hpt daemon",
			Action: func(ctx *cli.Context) error {
				svr, err := NewHPTServer(ctx.String("keystore"))
				if err != nil {
					return err
				}
				var lis net.Listener
				if os.Getenv("LISTEN_PID") == strconv.Itoa(os.Getpid()) {
					// We're a systemd socket-activated service.
					var err error
					f := os.NewFile(3, hptsock)
					lis, err = net.FileListener(f)
					if err != nil {
						return err
					}
				} else {
					// We are a regular, long-running daemon service.
					var err error
					port := ctx.String("port")
					lis, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
					if err != nil {
						return err
					}
				}
				return svr.Serve(lis)
			},
		},
	}

	// default action
	app.Action = func(ctx *cli.Context) error {
		args := ctx.Args()
		if !args.Present() {
			return errors.New("you must provide a state file")
		}
		if args.First() == "ssh-apply" {
			fmt.Println("did you mean \"apply-ssh\"?")
			os.Exit(1)
		}
		var paths = []string{args.First()}
		paths = append(paths, args.Tail()...)
		return run(paths...)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var (
	boldRed = color.New(color.FgRed).Add(color.Bold)
	red     = color.New(color.FgRed)
	white   = color.New(color.FgWhite)
	yellow  = color.New(color.FgYellow)
	blue    = color.New(color.FgBlue)
)

// run takes a sequence of paths to config files.
func run(paths ...string) error {

	// TODO support multiple configs
	path := paths[0]

	// colors don't work over ssh
	printState := func(s *ApplyState) {
		if s.Err != nil {
			red.Print(string(s.RenderShell()))
			return
		}
		blue.Println(string(s.RenderShell()))
	}
	conf, err := NewConfig(path)
	if err != nil {
		return err
	}
	ep, err := NewExecutionPlan(conf)
	if err != nil {
		return err
	}
	for {
		fn := ep.Next()
		if fn == nil {
			break
		}
		state := fn()
		printState(state)
	}

	return nil
}

// plan takes a sequence of paths to config files.
func plan(paths ...string) error {
	return nil
}
