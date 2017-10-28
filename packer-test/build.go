// This package is meant to be run with `go run test.go` to do a build/test.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/anxiousmodernman/easyssh"
	"github.com/digitalocean/godo"
	"github.com/fatih/color"

	sshlib "golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

var (
	region    = "sfo2"
	binary    = "hpt"
	testDir   = "packer-test"
	verbose   = true
	devImage  = "hpt-dev.json"
	baseImage = "centos7.json"
	sshUser   = "coleman"
	sshKey    = os.ExpandEnv("$HOME/.ssh/hackbox")
)

var usage = `
provide exactly one of: base, dev
`

// cli flags
var (
	keep = flag.Bool("keep", false, "whether to keep the vm online after the test is over")
)

func timer(then time.Time) func() {
	return func() {
		now := time.Now()
		yellow.Printf("run duration: %v seconds\n", now.Sub(then))
	}
}

func main() {
	flag.Parse()

	later := timer(time.Now())
	defer later()

	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	var image string
	switch os.Args[1] {
	case "dev":
		image = devImage
	case "base":
		image = baseImage
	default:
		fmt.Println(usage)
		os.Exit(1)
	}

	copyBinaryHere()

	cmd := exec.Command("packer", "build", image)
	stdout, _ := cmd.StdoutPipe()
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd start: %v", err)
	}

	re := `.*digitalocean.*ID: (?P<id>\d+)\).*`
	c1 := valueFromStream(re, stdout)
	cmd.Wait()

	tkn := token(os.Getenv("DIGITALOCEAN_API_TOKEN"))
	oauthClient := oauth2.NewClient(oauth2.NoContext, &tkn)
	client := godo.NewClient(oauthClient)

	dropletID := <-c1
	vm, err := launchServer(client, dropletID)
	if err != nil {
		later()
		log.Fatal(err)
	}
	addr, err := vm.PublicIPv4()
	if err != nil {
		later()
		log.Fatal(err)
	}

	// droplet is "ready" but we still need to give ssh some time.
	fmt.Println("sleeping for sshd...")
	time.Sleep(5 * time.Second)

	ssh := &easyssh.MakeConfig{
		User: sshUser, Server: addr, Key: sshKey,
		HostKeyCallback: sshlib.InsecureIgnoreHostKey(),
	}

	_ = func(cmd string) {
		output, err := ssh.Run(cmd)
		if err != nil {
			log.Fatalf("%s over ssh: %v", cmd, err)
		}
		fmt.Printf("%s: %v", cmd, output)
	}

	if *keep {
		// TODO don't do this
		fmt.Println("server is: ", addr)
		return
	}
	if err := destroyServer(client, addr); err != nil {
		later()
		log.Fatal(err)
	}
	return
}

func destroyServer(client *godo.Client, imageid string) error {
	id, err := strconv.Atoi(imageid)
	if err != nil {
		return err
	}

	resp, err := client.Droplets.Delete(context.TODO(), id)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return errors.New("please view DO console and clean up the vm manually")
	}
	return nil
}

// The token is our oauth2.TokenSource implementation.
type token string

// Token implements oauth2.TokenSource interface.
func (t *token) Token() (*oauth2.Token, error) {
	o2t := &oauth2.Token{AccessToken: string(*t)}
	return o2t, nil
}

// launchServer returns the IPv4 public IP or an error.
func launchServer(client *godo.Client, imageid string) (*godo.Droplet, error) {
	id, err := strconv.Atoi(imageid)
	if err != nil {
		return nil, err
	}
	fmt.Println("new image id:", id)

	ctx := context.TODO()

	var sshKeyID int
	keys, _, err := client.Keys.List(ctx, &godo.ListOptions{Page: 1, PerPage: 200})
	for _, k := range keys {
		if k.Name == "default" {
			sshKeyID = k.ID
		}
	}

	createRequest := &godo.DropletCreateRequest{
		Name:   fmt.Sprintf("hpt-test-%v", time.Now().Unix()),
		Region: region,
		Size:   "512mb",
		Image: godo.DropletCreateImage{
			ID: id,
		},
		IPv6: false,
		SSHKeys: []godo.DropletCreateSSHKey{
			godo.DropletCreateSSHKey{ID: sshKeyID},
		},
	}

	fmt.Println("creating vm with test image id")
	vm, _, err := client.Droplets.Create(ctx, createRequest)
	if err != nil {
		return nil, fmt.Errorf("droplet create: %v", err)
	}
	fmt.Println("waiting for vm to become ready")
	for vm.Status != "active" {
		fmt.Println("test droplet status:", vm.Status)
		time.Sleep(1 * time.Second)
		vm, _, err = client.Droplets.Get(ctx, vm.ID)
		if err != nil {
			return nil, fmt.Errorf("get droplet status: %v", err)
		}
	}
	return vm, nil
}

func valueFromStream(re string, r io.ReadCloser) chan string {
	var id string
	reg, scnr, done := regexp.MustCompile(re), bufio.NewScanner(r), make(chan string)
	go func() {
		for scnr.Scan() {
			line := scnr.Text()
			fmt.Println(line) // this is what gives us output
			if match := reg.FindStringSubmatch(line); match != nil {
				paramsMap := make(map[string]string)
				for i, name := range reg.SubexpNames() {
					if i > 0 && i <= len(match) {
						paramsMap[name] = match[i]
					}
				}
				id = paramsMap["id"]
			}
		}
		done <- id
	}()
	return done
}

func runCommand(cmd string, args ...string) {
	c := exec.Command(cmd, args...)
	out, err := c.CombinedOutput()
	if err != nil {
		log.Fatalf("exec %v %v: %v", cmd, args, err)
	}
	fmt.Println(string(out))
}

func copyBinaryHere() {
	pwd, _ := os.Getwd()
	cd := func() { os.Chdir(pwd) }

	// do go build
	os.Chdir("..")
	os.Setenv("GOOS", "linux")
	os.Setenv("GOARCH", "amd64")
	runCommand("go", "build", "-o", binary)
	runCommand("mv", binary, testDir)
	cd()
}

var (
	boldRed = color.New(color.FgRed).Add(color.Bold)
	white   = color.New(color.FgWhite)
	yellow  = color.New(color.FgYellow)
)
