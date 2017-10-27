package main

import (
	"fmt"
	"io"
	"os"
	"regexp"

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
			// env vars will be different based on provider. What do we do?
			bc, err := NewBucketClient(v.URL,
				os.Getenv("DO_ACCESS_KEY"), os.Getenv("DO_SECRET_ACCESS_KEY"))
			if err != nil {
				return nil, err
			}
			// pass the bucket name here
			return NewBucketResolver(v.Name, bc), nil
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

// NewBucketResolver ...
func NewBucketResolver(bucket string, bc *BucketClient) *BucketResolver {
	return &BucketResolver{bc, bucket}
}

// BucketResolver wraps our BucketClient as a resolver.
type BucketResolver struct {
	*BucketClient
	bucket string
}

// Get implements Resolver for an S3-like bucket.
func (br *BucketResolver) Get(path string) (io.Reader, error) {
	return br.conn.GetObject(br.bucket, path)
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

func ParseResolverPath(path string) (string, string) {
	m := getParams("^(?P<resolver>.*)+[:]{1}[/]{2}(?P<path>.*)+", path)
	resolver, parsedPath := m["resolver"], m["path"]
	if resolver == "" && parsedPath == "" {
		// not a protocol: return original path, and
		// caller can treat empty resolver as a signal
		// that it's a local path
		return "", path
	}
	return resolver, parsedPath
}

func getParams(regEx, path string) map[string]string {

	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(path)

	m := make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			m[name] = match[i]
		}
	}
	return m
}
