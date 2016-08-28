package storage

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"google.golang.org/api/option"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"

	"cloud.google.com/go/storage"
)

// gs://bucket[/prefix]

type (
	GoogleStorageProvider byte
	googleCloudStorageRef struct {
		clientID     string
		clientSecret string
	}
)

func init() {
	//Register("gs", GoogleStorageProvider)
}

func (gsp GoogleStorageProvider) auth() (*jwt.Config, error) {
	var key []byte
	if authFilePath := os.Getenv("GCLOUD_KEY_FILE"); authFilePath != "" {
		if bs, err := ioutil.ReadFile(authFilePath); err == nil {
			key = bs
		}
	} else if auth := os.Getenv("GCLOUD_KEY"); auth != "" {
		key = []byte(auth)
	}
	return google.JWTConfigFromJSON(
		key,
		storage.ScopeReadOnly,
	)
}

func (gsp GoogleStorageProvider) Get(rawurl string) (io.ReadCloser, error) {
	conf, err := gsp.auth()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithTokenSource(conf.TokenSource(ctx)))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	return nil, fmt.Errorf("not implemented")
}

func (gsp GoogleStorageProvider) Put(rawurl string, rdr io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (gsp GoogleStorageProvider) List(rawurl string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
