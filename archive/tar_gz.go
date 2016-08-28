package archive

import (
	"compress/gzip"
	"io"
)

// GZip extracts gzip tar archives
var GZip gzipArchiveProvider

type gzipArchiveProvider struct{}

func (p gzipArchiveProvider) ExtractReader(dst string, src io.Reader) error {
	z, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer z.Close()
	return Tar.ExtractReader(dst, z)
}
