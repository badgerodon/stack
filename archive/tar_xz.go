// +build cgo

package archive

import (
	"io"
	"os"

	"github.com/remyoudompheng/go-liblzma"
)

type XZArchiveProvider struct{}

var XZ = &XZArchiveProvider{}

func init() {
	Register(".tar.xz", XZ)
	Register(".tar.lz", XZ)
	Register(".tar.lzma", XZ)
	Register(".txz", XZ)
	Register(".tlz", XZ)
}

func (txz *XZArchiveProvider) Extract(dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	return txz.ExtractReader(dst, f)
}

func (txz *XZArchiveProvider) ExtractReader(dst string, src io.Reader) error {
	z, err := xz.NewReader(src)
	if err != nil {
		return err
	}
	return Tar.ExtractReader(dst, z)
}
