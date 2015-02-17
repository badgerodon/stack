package sync

import (
	"io"
	"os"
	"strconv"
	"time"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
)

type (
	S3Source struct {
		auth         aws.Auth
		region       aws.Region
		bucket, path string
	}
)

var S3PollingInterval time.Duration

func init() {
	interval, _ := strconv.Atoi(os.Getenv("POLLING_INTERVAL"))
	if interval <= 0 {
		interval = 15
	}
	S3PollingInterval = time.Second * time.Duration(interval)
}

func (ss S3Source) Watch() *Watcher {
	return NewWatcher(func(done <-chan struct{}, change chan<- struct{}) {
		ticker := time.NewTicker(S3PollingInterval)
		defer ticker.Stop()

		changed := false
		lastETag := ""

		for {
			// if we've changed, signal the watcher
			if changed {
				select {
				case <-done:
					return
				// this gives us a chance to cleanup if no one ever listens
				//   after 15 seconds we'll just loop back around and contine waiting
				case <-ticker.C:
				case change <- struct{}{}:
					changed = false
				}
			} else {
				select {
				case <-done:
					return
				case <-ticker.C:
					res, err := s3.New(ss.auth, ss.region).Bucket(ss.bucket).Head(ss.path, nil)
					if err == nil {
						if res.Header.Get("ETag") != lastETag {
							lastETag = res.Header.Get("ETag")
							changed = true
						}
					}
				}
			}
		}
	})
}

func (ss S3Source) Sync(dst string) error {
	rc, err := s3.New(ss.auth, ss.region).Bucket(ss.bucket).GetReader(ss.path)
	if err != nil {
		return err
	}
	defer rc.Close()

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, rc)
	if err != nil {
		return err
	}

	return nil
}

func (s3 S3Source) Ext() string {
	return ""
}
