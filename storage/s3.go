package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
)

type (
	S3Provider struct {
	}

	s3ref struct {
		auth   aws.Auth
		region aws.Region
		bucket string
		path   string
	}
)

func (s3p *S3Provider) parse(rawurl string) s3ref {
	ref := s3ref{path: rawurl, region: aws.USEast}
	u, err := url.Parse(rawurl)
	if err == nil {
		for _, region := range aws.Regions {
			if strings.HasPrefix(region.S3Endpoint, "https://"+u.Host) {
				ref.region = region
			}
		}
	} else {

	}
	fmt.Println(ref)
	return ref
}

func (s3p *S3Provider) Get(rawurl string) (io.ReadCloser, error) {
	ref := s3p.parse(rawurl)
	return s3.New(ref.auth, ref.region).Bucket(ref.bucket).GetReader(ref.path)
}

func (s3p *S3Provider) Put(rawurl string, rdr io.Reader) error {
	ref := s3p.parse(rawurl)
	contentType := mime.TypeByExtension(filepath.Ext(ref.path))
	bucket := s3.New(ref.auth, ref.region).Bucket(ref.bucket)

	if sz, err := getSize(rdr); err == nil {
		return bucket.PutReader(ref.path, rdr, sz, contentType, "", s3.Options{})
	}

	bs, err := ioutil.ReadAll(rdr)
	if err != nil {
		return err
	}

	return bucket.Put(ref.path, bs, contentType, "", s3.Options{})
}

func (s3p *S3Provider) List(rawurl string) ([]string, error) {
	ref := s3p.parse(rawurl)
	cl := s3.New(ref.auth, ref.region).Bucket(ref.bucket)
	marker := ""
	truncated := true
	names := make([]string, 0)
	for truncated {
		res, err := cl.List(rawurl, "/", marker, 1000)
		if err != nil {
			return nil, err
		}
		for _, key := range res.Contents {
			names = append(names, key.Key)
		}
		marker = res.NextMarker
		truncated = res.IsTruncated
	}
	return names, nil
}
