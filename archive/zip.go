package archive

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// Zip extracts .zip files
var Zip zipExtractor

type zipExtractor struct{}

func (z *zipExtractor) Extract(dst, src string) error {
	rc, err := zip.OpenReader(src)
	if err != nil {
		return err
	}

	for _, zf := range rc.File {
		fp := filepath.Join(dst, zf.Name)
		if zf.FileInfo().IsDir() {
			os.MkdirAll(fp, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(fp), 0755)

		zfr, err := zf.Open()
		if err != nil {
			return err
		}

		f, err := os.Create(fp)
		if err != nil {
			zfr.Close()
			return err
		}
		_, err = io.Copy(f, zfr)
		f.Close()
		zfr.Close()
		if err != nil {
			return err
		}
		os.Chmod(fp, zf.Mode())
	}

	return nil
}
