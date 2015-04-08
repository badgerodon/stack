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

func TestXZ(t *testing.T) {
	assert := assert.New(t)
	rawStrings := []string{
		"/Td6WFoAAATm1rRGAgAhARYAAAB0L+Wj4Cf/AIZdADQZSe6N8LrI/5v/8gxplkSzD8Cf2p6mVVyx9h+vPawsxCPKSvcdTQr0ZK/WC0r7L6YB+YMiDPBMzTgUsYbDk847aDpq7x8CzeBQLZQ4SaF5EzXUOMU0BF21S+JYUuAnPcP4spkpRiAQIh+p2+RsWO4x9MxKlLnaS8jLqxZAZ1cuUQdvjaJgAAAAXWjacOjBQGIAAaIBgFAAAAPVBCuxxGf7AgAAAAAEWVo=",
		"XQAAgAD//////////wA0GUnujfC6yP+b//IMaZZEsw/An9qeplVcsfYfrz2sLMQjykr3HU0K9GSv1gtK+y+mAfmDIgzwTM04FLGGw5POO2g6au8fAs3gUC2UOEmheRM11DjFNARdtUviWFLgJz3D+LKZKUYgECIfqdvkbFjuMfTMSpS52kvIy6sWQGdXLlEHcIVr23z/sUVdAA==",
	}

	for _, rawString := range rawStrings {
		raw, _ := base64.StdEncoding.DecodeString(rawString)
		folder := filepath.Join(os.TempDir(), uuid.NewRandom().String())
		os.MkdirAll(folder, 0777)
		defer os.RemoveAll(folder)

		err := XZ.ExtractReader(folder, bytes.NewReader(raw))
		assert.Nil(err)

		bs, err := ioutil.ReadFile(filepath.Join(folder, "hello.txt"))
		assert.Nil(err)
		assert.Equal("Hello World\n", string(bs))
	}
}
