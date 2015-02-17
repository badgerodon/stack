package sync

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/goamz/goamz/aws"
	"github.com/stretchr/testify/assert"
)

func TestS3Sync(t *testing.T) {
	assert := assert.New(t)

	li, err := net.Listen("tcp", "127.0.0.1:0")
	assert.Nil(err)
	defer li.Close()

	go http.Serve(li, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		assert.Equal(req.Method, "GET")
		assert.Equal(req.URL.Path, "/FAKE_BUCKET/FAKE/PATH.EXT")
		res.Write([]byte("HELLO"))
	}))

	region := aws.Region{
		Name:       "faux-region-1",
		S3Endpoint: "http://" + li.Addr().String() + "/",
	}

	src := &S3Source{
		region: region,
		bucket: "FAKE_BUCKET",
		path:   "FAKE/PATH.EXT",
	}

	dst := filepath.Join(os.TempDir(), uuid.New())
	defer os.Remove(dst)

	err = src.Sync(dst)
	assert.Nil(err)

	bs, err := ioutil.ReadFile(dst)
	assert.Nil(err)
	assert.Equal("HELLO", string(bs))
}

func TestS3Watch(t *testing.T) {
	assert := assert.New(t)

	li, err := net.Listen("tcp", "127.0.0.1:0")
	assert.Nil(err)
	defer li.Close()

	var mu sync.Mutex
	var etag = "1234"

	go http.Serve(li, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		assert.Equal(req.Method, "HEAD")
		assert.Equal(req.URL.Path, "/FAKE_BUCKET/FAKE/PATH.EXT")
		mu.Lock()
		res.Header().Set("ETag", etag)
		mu.Unlock()
	}))

	region := aws.Region{
		Name:       "faux-region-1",
		S3Endpoint: "http://" + li.Addr().String() + "/",
	}

	src := &S3Source{
		region: region,
		bucket: "FAKE_BUCKET",
		path:   "FAKE/PATH.EXT",
	}

	S3PollingInterval = time.Millisecond * 10
	watcher := src.Watch()
	changed := false
	select {
	case <-watcher.C:
		changed = true
	case <-time.After(time.Millisecond * 20):
	}
	assert.True(changed, "should start changed")
	changed = false
	select {
	case <-watcher.C:
		changed = true
	case <-time.After(time.Millisecond * 20):
	}
	assert.False(changed, "change should not be detected")

	mu.Lock()
	etag = "5678"
	mu.Unlock()

	changed = false
	select {
	case <-watcher.C:
		changed = true
	case <-time.After(time.Millisecond * 20):
	}
	assert.True(changed, "should change with new etag")
	changed = false
	select {
	case <-watcher.C:
		changed = true
	case <-time.After(time.Millisecond * 20):
	}
	assert.False(changed, "change should not be detected")
}
