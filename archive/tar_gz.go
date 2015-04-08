package archive

import (
	"compress/gzip"
	"io"
	"os"
)

type GZipArchiveProvider struct{}

var GZip = &GZipArchiveProvider{}

func init() {
	Register(".tar.gz", GZip)
	Register(".tgz", GZip)
}

func (tgz *GZipArchiveProvider) Extract(dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	return tgz.ExtractReader(dst, f)
}

func (tgz *GZipArchiveProvider) ExtractReader(dst string, src io.Reader) error {
	z, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer z.Close()
	return Tar.ExtractReader(dst, z)
}
