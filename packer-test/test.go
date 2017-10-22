// This package is meant to be run with `go run test.go` to do a build/test.
package main

import (
	"bufio"
	"context"
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

	"golang.org/x/oauth2"
)

var (
	binary  = "hpt"
	testDir = "packer-test"
	verbose = true
)

// we're a script
func main() {
	then := time.Now()
	defer func() {
		now := time.Now()
		fmt.Printf("test duration: %v seconds\n", now.Sub(then))
	}()

	runCommand := func(cmd string, args ...string) {
		c := exec.Command(cmd, args...)
		out, err := c.CombinedOutput()
		if err != nil {
			now := time.Now()
			fmt.Printf("test duration: %v seconds\n", now.Sub(then))
			log.Fatalf("exec %v %v: %v", cmd, args, err)
		}
		fmt.Println(string(out))
	}

	pwd, _ := os.Getwd()
	cd := func() { os.Chdir(pwd) }

	// do go build
	os.Chdir("..")
	os.Setenv("GOOS", "linux")
	os.Setenv("GOARCH", "amd64")
	runCommand("go", "build", "-o", binary)
	runCommand("mv", binary, testDir)
	cd()

	// do the packer build
	cmd := exec.Command("packer", "build", "centos7.json")

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	defer stdout.Close()
	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd start: %v", err)
	}

	valueFromStream := func(re string, r io.ReadCloser) chan string {
		var id string
		reg, scnr, done := regexp.MustCompile(re), bufio.NewScanner(r), make(chan string)
		go func() {
			for scnr.Scan() {
				line := scnr.Text()
				fmt.Println(line)
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

	var re = `.*digitalocean.*ID: (?P<id>\d+)\).*`
	c1 := valueFromStream(re, stdout)

	cmd.Wait()
	idstr := <-c1

	id, err := strconv.Atoi(idstr)
	if err != nil {
		log.Fatalf("could not parse image id as int: %v", err)
	}

	fmt.Println("captured ids")
	fmt.Println("stdout", id)

	// string -> token type conversion
	tkn := token(os.Getenv("DIGITALOCEAN_API_TOKEN"))

	// Now it's DO API time.

	oauthClient := oauth2.NewClient(oauth2.NoContext, &tkn)
	client := godo.NewClient(oauthClient)
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
		Region: "nyc3",
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
		log.Fatalf("droplet create: %v", err)
	}
	fmt.Println("waiting for vm to become ready")
	for vm.Status != "active" {
		fmt.Println("test droplet status:", vm.Status)
		time.Sleep(1 * time.Second)
		vm, _, err = client.Droplets.Get(ctx, vm.ID)
		if err != nil {
			log.Fatalf("get droplet status: %v", err)
		}
	}
	// droplet is "ready" but we still need to give ssh some time.
	fmt.Println("sleeping for sshd...")
	time.Sleep(5 * time.Second)
	addr, _ := vm.PublicIPv4()
	ssh := &easyssh.MakeConfig{User: "coleman", Server: addr, Key: "/home/coleman/.ssh/hackbox"}

	runSSH := func(cmd string) {
		output, err := ssh.Run(cmd)
		if err != nil {
			log.Fatalf("%s over ssh: %v", cmd, err)
		}
		fmt.Printf("%s: %v", cmd, output)
	}

	fmt.Println("running some commands remotely...")
	// Tests/Verifications over ssh
	runSSH("groups")
	runSSH("sudo systemctl status sshd")
	fmt.Println("sudo commands successful!")

	if true {
		// TODO don't do this
		fmt.Println("server is: ", addr)
		return
	}

	resp, err := client.Droplets.Delete(ctx, vm.ID)
	if err != nil {
		log.Fatalf("error cleaning up test vm: %v", err)
		log.Println("please view DO console and clean up the vm manually")
	}
	if resp.StatusCode != 204 {
		log.Fatalf("error cleaning up test vm: %v", err)
		log.Println("please view DO console and clean up the vm manually")
	}

}

// The token is our oauth2.TokenSource implementation.
type token string

// Token implements oauth2.TokenSource interface.
func (t *token) Token() (*oauth2.Token, error) {
	o2t := &oauth2.Token{AccessToken: string(*t)}
	return o2t, nil
}
