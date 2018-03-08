package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/anxiousmodernman/hpt/proto/server"
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
		Name:   "keystore",
		Usage:  "Path to keystore for a target's hpt instance",
		EnvVar: "HPT_KEYSTORE",
		Value:  "/etc/hpt/keystore.db",
	}
	keystoreOutputPath := cli.StringFlag{
		Name:  "output",
		Usage: "output path for generated keystores",
		Value: "keystore.db",
	}
	serverPort := cli.StringFlag{
		Name:  "port",
		Usage: "Port to listen on",
		Value: "6632", // GGEZ
	}
	targetName := cli.StringFlag{
		Name:  "target",
		Usage: "name of target; identifies target keypairs in our local keystore",
	}
	targetIP := cli.StringFlag{
		Name:  "ip",
		Usage: "ip of target; must be paired with --target for remote execution",
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
			Name:  "list-target-keys",
			Flags: []cli.Flag{keystorePath, targetName},
			Usage: "generate a boltdb keystore to be copied to a target server",
			Action: func(ctx *cli.Context) error {
				ks := ctx.String("keystore")
				return ListTargetServerKeys(ks)

			},
		},
		cli.Command{
			Name:  "gen-target-keys",
			Flags: []cli.Flag{keystorePath, targetName, keystoreOutputPath},
			Usage: "print the named targets and their public keys",
			Action: func(ctx *cli.Context) error {
				t, ks := ctx.String("target"), ctx.String("keystore")
				data, err := GenerateTargetServerKeys(t, ks)
				if err != nil {
					return err
				}
				out := ctx.String("output")
				return ioutil.WriteFile(out, data, 0744)
			},
		},
		cli.Command{
			Name:  "gen-client-keys",
			Flags: []cli.Flag{keystorePath},
			Usage: "generate client keys for this instance; do this once; WARNING: invalidates target keystores",
			Action: func(ctx *cli.Context) error {
				ks := ctx.String("keystore")
				return GenerateClientKeys(ks)
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
				_, svr, err := NewHPTServer(ctx.String("keystore"))
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
					lis, err = net.Listen("tcp4", fmt.Sprintf("0.0.0.0:%s", port))
					if err != nil {
						return err
					}
				}

				return svr.Serve(lis)
			},
		},
	}

	// global flags
	app.Flags = []cli.Flag{targetName, targetIP, serverPort, keystorePath}

	// default action
	app.Action = func(ctx *cli.Context) error {
		args := ctx.Args()
		if !args.Present() {
			return errors.New("you must provide a state file")
		}
		target := ctx.String("target")
		ip := ctx.String("ip")
		port := ctx.String("port")
		ks := ctx.String("keystore")
		if target == "" {
			// no target specified, assuming local execution
			var paths = []string{args.First()}
			paths = append(paths, args.Tail()...)
			return run(paths...)
		}
		if ip == "" {
			return errors.New("if --target specified, must specify --ip, too")
		}
		addr := fmt.Sprintf("%s:%s", ip, port)
		c, err := NewHPTClient(ks, target, addr)
		if err != nil {
			return err
		}
		// read config passed in
		var paths = []string{args.First()}
		paths = append(paths, args.Tail()...)
		// only support single conf right now:
		singleConf := paths[0]
		data, err := ioutil.ReadFile(singleConf)
		if err != nil {
			return err
		}
		req := server.Config{data}
		stream, err := c.Apply(context.TODO(), &req)
		if err != nil {
			return err
		}
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					// normal termination
					break
				}
				return err
			}

			_ = msg
			// switch msg.(type) {
			// case *server.ApplyResult_Metadata:
			// 	md, _ := msg.(*server.ApplyResult_Metadata)
			// 	blue.Printf("Name: %s", md.Metadata.Name)
			// 	blue.Printf("Outcome: %s", md.Metadata.Result.String())
			// }
		}

		// send to server
		return nil
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
