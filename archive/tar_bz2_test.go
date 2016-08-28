package archive

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestBZip2(t *testing.T) {
	assert := assert.New(t)
	// created with: tar -cj
	raw, _ := base64.StdEncoding.DecodeString("QlpoOTFBWSZTWcwUjO4AAIT/kNIQAYBAAX+AAEAAgH5EnsAEAAAIIAByGlBkGEA0ZDJ5QZBEanpDQ9TExpMOc8NrpwIIErCQSPbgVYgERAzUzFEo1q1ErQDD0gyRgjqZUbB/jnFMbl7hjGFjtzPsmv2schs2qn5bGlYuAhZ0aIae6bERAfi7kinChIZgpGdw")

	folder := filepath.Join(os.TempDir(), uuid.NewV4().String())
	os.MkdirAll(folder, 0777)
	defer os.RemoveAll(folder)

	err := BZip2.ExtractReader(folder, bytes.NewReader(raw))
	assert.Nil(err)

	bs, err := ioutil.ReadFile(filepath.Join(folder, "hello.txt"))
	assert.Nil(err)
	assert.Equal("Hello World\n", string(bs))
}
