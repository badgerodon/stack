package archive

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
)

type TarArchiveProvider struct{}

var Tar = &TarArchiveProvider{}

func init() {
	Register(".tar", Tar)
}

func (t *TarArchiveProvider) Create(dst, src string) error {
	panic("not implemented")
}

func (t *TarArchiveProvider) Extract(dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.ExtractReader(dst, f)
}

func (t *TarArchiveProvider) ExtractReader(dst string, src io.Reader) error {
	tr := tar.NewReader(src)

	for {
		var h *tar.Header
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fp := filepath.Join(dst, h.Name)
		if h.FileInfo().IsDir() {
			os.MkdirAll(fp, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(fp), 0755)

		f, err := os.Create(fp)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, tr)
		f.Close()
		if err != nil {
			return err
		}
		os.Chmod(fp, os.FileMode(h.Mode))
	}

	return nil
}
