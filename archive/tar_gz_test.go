package archive

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"code.google.com/p/go-uuid/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGZip(t *testing.T) {
	assert := assert.New(t)
	raw, _ := base64.StdEncoding.DecodeString("H4sIAPFDJVUAA+3RMQ7CMAyF4c6cwidAcZq6V+AGzAUidYhUqQTB8UkrgcRSMRCx/N/yBnt4lseY0rTPj9zU45yzEGTJ3ro1i1cWGkR9p+pb672J01a9NeIqdnq7XfMwlyrnIcXTxt59jDFtzD+Pkh+3rOaw/F+O05wuu393AQAAAAAAAAAAAAAAAAB87wlzTKkTACgAAA==")

	folder := filepath.Join(os.TempDir(), uuid.NewRandom().String())
	os.MkdirAll(folder, 0777)
	defer os.RemoveAll(folder)

	err := GZip.ExtractReader(folder, bytes.NewReader(raw))
	assert.Nil(err)

	bs, err := ioutil.ReadFile(filepath.Join(folder, "hello.txt"))
	assert.Nil(err)
	assert.Equal("Hello World\n", string(bs))
}
