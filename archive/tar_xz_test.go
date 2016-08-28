package archive

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/badgerodon/stack/archive/testdata"
	"github.com/satori/go.uuid"
)

func TestXZ(t *testing.T) {
	bs, _ := base64.StdEncoding.DecodeString(testdata.XZTar)
	folder := filepath.Join(os.TempDir(), uuid.NewV4().String())
	os.MkdirAll(folder, 0777)
	defer os.RemoveAll(folder)

	err := XZ.ExtractReader(folder, bytes.NewReader(bs))
	if err != nil {
		t.Fatalf("failed to extract archive: %v", err)
	}

	bs, err = ioutil.ReadFile(filepath.Join(folder, "hello.txt"))
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if testdata.Original != string(bs) {
		t.Fatalf("expected `%v`, got `%v`", testdata.Original, string(bs))
	}
}

func TestLZMA(t *testing.T) {
	bs, _ := base64.StdEncoding.DecodeString(testdata.LZMATar)
	folder := filepath.Join(os.TempDir(), uuid.NewV4().String())
	os.MkdirAll(folder, 0777)
	defer os.RemoveAll(folder)

	err := LZMA.ExtractReader(folder, bytes.NewReader(bs))
	if err != nil {
		t.Fatalf("failed to extract archive: %v", err)
	}

	bs, err = ioutil.ReadFile(filepath.Join(folder, "hello.txt"))
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if testdata.Original != string(bs) {
		t.Fatalf("expected `%v`, got `%v`", testdata.Original, string(bs))
	}
}
