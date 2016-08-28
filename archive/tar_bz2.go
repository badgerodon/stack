package archive

import (
	"compress/bzip2"
	"io"
)

// BZip2 extracts bzip2 tar archives
var BZip2 bzip2Provider

type bzip2Provider struct{}

func (p bzip2Provider) ExtractReader(dst string, src io.Reader) error {
	bz := bzip2.NewReader(src)
	return Tar.ExtractReader(dst, bz)
}
