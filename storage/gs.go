package storage

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// gs://bucket[/prefix]

type (
	googleStorage         struct{}
	googleCloudStorageRef struct {
		clientID     string
		clientSecret string
	}
)

func (s googleStorage) client(loc Location, scope string) (*storage.Client, error) {
	var key []byte
	if authFilePath := os.Getenv("GCLOUD_KEY_FILE"); authFilePath != "" {
		if bs, err := ioutil.ReadFile(authFilePath); err == nil {
			key = bs
		}
	} else if auth := os.Getenv("GCLOUD_KEY"); auth != "" {
		key = []byte(auth)
	}

	ctx := context.Background()
	if tok, err := google.JWTConfigFromJSON(key, scope); err == nil {
		return storage.NewClient(ctx, option.WithTokenSource(tok.TokenSource(ctx)))
	}
	// fallback to default application credentials
	tok, err := google.DefaultTokenSource(ctx, scope)
	if err != nil {
		return nil, err
	}
	return storage.NewClient(ctx, option.WithTokenSource(tok))
}

func (s googleStorage) Get(loc Location) (io.ReadCloser, error) {
	client, err := s.client(loc, storage.ScopeReadOnly)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	path := loc.Path()
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	bucket := client.Bucket(loc.Host())
	object := bucket.Object(path)

	return object.NewReader(context.Background())
}

func (s googleStorage) List(loc Location) ([]string, error) {
	client, err := s.client(loc, storage.ScopeReadOnly)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	path := loc.Path()
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	bucket := client.Bucket(loc.Host())
	it := bucket.Objects(context.Background(), &storage.Query{
		Prefix: path,
	})
	var names []string
	for {
		object, err := it.Next()
		if err == storage.Done {
			break
		} else if err != nil {
			return nil, err
		}
		names = append(names, object.Name[len(path):])
	}
	return names, nil
}

func (s googleStorage) Put(loc Location, src io.Reader) error {
	client, err := s.client(loc, storage.ScopeReadOnly)
	if err != nil {
		return err
	}
	defer client.Close()

	path := loc.Path()
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	bucket := client.Bucket(loc.Host())
	w := bucket.Object(path).NewWriter(context.Background())
	_, err = io.Copy(w, src)
	if err != nil {
		w.Close()
		return err
	}
	return w.Close()
}
