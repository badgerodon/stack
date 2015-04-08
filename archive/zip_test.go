package archive

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"code.google.com/p/go-uuid/uuid"
	"github.com/stretchr/testify/assert"
)

func TestZip(t *testing.T) {
	assert := assert.New(t)
	raw, _ := base64.StdEncoding.DecodeString("UEsDBAoAAAAAAAtNiEbj5ZWwDAAAAAwAAAAJABwAaGVsbG8udHh0VVQJAAPWPSVVV0QlVXV4CwABBPUBAAAEAAAAAEhlbGxvIFdvcmxkClBLAQIeAwoAAAAAAAtNiEbj5ZWwDAAAAAwAAAAJABgAAAAAAAEAAACkgQAAAABoZWxsby50eHRVVAUAA9Y9JVV1eAsAAQT1AQAABAAAAABQSwUGAAAAAAEAAQBPAAAATwAAAAAA")

	fn := filepath.Join(os.TempDir(), uuid.NewRandom().String())
	defer os.Remove(fn)

	ioutil.WriteFile(fn, raw, 0777)

	folder := filepath.Join(os.TempDir(), uuid.NewRandom().String())
	os.MkdirAll(folder, 0777)
	defer os.RemoveAll(folder)

	err := Zip.Extract(folder, fn)
	assert.Nil(err)

	bs, err := ioutil.ReadFile(filepath.Join(folder, "hello.txt"))
	assert.Nil(err)
	assert.Equal("Hello World\n", string(bs))
}
