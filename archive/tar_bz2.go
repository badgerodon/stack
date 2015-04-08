package archive

import (
	"compress/bzip2"
	"io"
	"os"
)

type BZip2Provider struct{}

var BZip2 = &BZip2Provider{}

func init() {
	Register(".tar.bz2", BZip2)
	Register(".tbz", BZip2)
	Register(".tbz2", BZip2)
	Register(".tb2", BZip2)
}

func (tbz *BZip2Provider) Extract(dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	return tbz.ExtractReader(dst, f)
}

func (tbz *BZip2Provider) ExtractReader(dst string, src io.Reader) error {
	bz := bzip2.NewReader(src)
	return Tar.ExtractReader(dst, bz)
}
