package storage

import (
	"testing"

	"github.com/goamz/goamz/aws"
	"github.com/stretchr/testify/assert"
)

func TestS3Parse(t *testing.T) {
	assert := assert.New(t)
	var s3p S3Provider
	ref := s3p.parse("s3://s3.cn-north-1.amazonaws.com.cn/some-bucket/some-path/to/a/file")
	assert.Equal(ref.region, aws.CNNorth)
}
