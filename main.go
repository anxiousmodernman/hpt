package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/minio/minio-go"
)

var (
	confPath = flag.String("conf", "", "path to config file")
)

type Path string

// automation for small infrastructures
func main() {
	// same binary for exec and serve
	// make it work for nice image builds first

	flag.Parse()
	accessKey := os.Getenv("DO_ACCESS_KEY")
	secKey := os.Getenv("DO_SECRET_ACCESS_KEY")
	ssl := true

	// fetch your own config
	conf, err := NewConfig(*confPath)
	if err != nil {
		log.Fatal(err)
	}
	_ = conf

	// Initiate a client using DigitalOcean Spaces.
	client, err := minio.New("nyc3.digitaloceanspaces.com", accessKey, secKey, ssl)
	if err != nil {
		log.Fatal(err)
	}

	opts := minio.GetObjectOptions{}
	obj, err := client.GetObject("coleman", "ssh-keys/hackbox", opts)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(obj)
	if err := ioutil.WriteFile("hackbox", data, 0600); err != nil {
		log.Fatal(err)
	}
}
