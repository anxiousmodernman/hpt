package main

import (
	"fmt"
	"io"
	"os"

	minio "github.com/minio/minio-go"
)

// Resolver gives us an interface backed by one of our concrete client
// implementations. Given a string like "foo://some/path.txt" in our config,
// a corresponding named resolver can be located in our config, such as
// an S3-like client declared elsewhere in TOML like:
//
//  [bucket.foo]
//  url = "nyc3.digitaloceanspaces.com"
//  name = "bucketname"
//
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
			// TODO extract secrets into executor config
			bc, err := NewBucketClient(v.URL,
				os.Getenv("DO_ACCESS_KEY"), os.Getenv("DO_SECRET_ACCESS_KEY"))
			if err != nil {
				return nil, err
			}
			return NewBucketResolver(name, bc), nil
		}
	}
	if !found {
		return nil, fmt.Errorf("unknown resolver: %s", name)
	}

	// This should never happen. Enjoy the nil reference!
	return nil, nil
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

// NewBucketClient ...
func NewBucketClient(url, accessKey, secretKey string) (*BucketClient, error) {
	var bc BucketClient
	// ssl is default true, for now
	c, err := minio.New(url, accessKey, secretKey, true)
	if err != nil {
		return nil, err
	}
	bc.conn = c
	return &bc, nil
}
