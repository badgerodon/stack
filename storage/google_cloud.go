package storage

import (
	"fmt"
	"io"
)

// gs://bucket[/prefix]

type (
	GoogleCloudStorageProvider byte
	googleCloudStorageRef      struct {
		clientID     string
		clientSecret string
	}
)

const (
	DefaultGoogleCloudStorageClientID     = "304359942533-ra5badnhb5f1umi5vj4p5oohfhdiq8v8.apps.googleusercontent.com"
	DefaultGoogleCloudStorageClientSecret = "2ORaxB_WysnMlfeYW5yZsBgH"
)

func (gcsp GoogleCloudStorageProvider) Get(rawurl string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (gcsp GoogleCloudStorageProvider) Put(rawurl string, rdr io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (gcsp GoogleCloudStorageProvider) List(rawurl string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
