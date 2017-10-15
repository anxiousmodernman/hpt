package main

import (
	"flag"
	"fmt"
	"io"
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

// NewBucketClient ...
func NewBucketClient(url, accessKey, secretKey string) (*BucketClient, error) {
	opts := minio.GetObjectOptions{}
	var bc BucketClient
	// ssl is default true, for now
	c, err := minio.New(url, accessKey, secretKey, true)
	if err != nil {
		return nil, err
	}
	bc.conn = c
	return &bc, nil
}

// BucketClient ...
type BucketClient struct {
	conn *minio.Client
}

func NewBucketResolver(bucket string, bc *BucketClient) *BucketResolver {
	return &BucketResolver{bc, bucket}
}

type BucketResolver struct {
	*BucketClient
	bucket string
}

func (br *BucketResolver) Get(path string) (io.Reader, error) {
	opts := minio.GetObjectOptions{}
	return br.conn.GetObject(br.bucket, path, opts)
}

// Resolver ...
type Resolver interface {
	// Is a nil reader meaningful? (e.g., missing, not error?)
	// This is a boltdb query's behavior, but maybe not everybody's.
	// need to analyze behavior of real client impls to make this true.
	Get(path string) (io.Reader, error)
}

// BuildResolver takes a resolver identifier from a source string, instantiates
// the appropriate client, and returns an interface that can resolve the value.
func BuildResolver(name string, conf Config) (Resolver, error) {
	var found bool
	for k, v := range conf.Buckets {
		if k == name {
			bc, err := NewBucketClient(v.URL,
				os.Getenv("DO_ACCESS_KEY"), os.Getenv("DO_SECRET_ACCESS_KEY"))
			return NewBucketResolver(name, bc), nil
		}
	}
	if !found {
		return nil, fmt.Errorf("unknown resolver: %s", name)
	}

	// This should never happen. Enjoy the nil reference!
	return nil, nil
}
