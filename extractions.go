package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func extractTar(dst string, rdr io.Reader) error {
	tr := tar.NewReader(rdr)

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
		os.MkdirAll(fp, 0755)
		if h.FileInfo().IsDir() {
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

func extractTarGz(dst string, rdr *os.File) error {
	zrdr, err := gzip.NewReader(rdr)
	if err != nil {
		return fmt.Errorf("invalid gzip file %v: %v", rdr.Name(), err)
	}
	defer zrdr.Close()

	return extractTar(dst, zrdr)
}

func extractZip(dst string, rdr *os.File) error {
	fi, err := rdr.Stat()
	if err != nil {
		return err
	}
	z, err := zip.NewReader(rdr, fi.Size())
	if err != nil {
		return fmt.Errorf("invalid zip file: %v: %v", rdr.Name())
	}

	for _, zf := range z.File {
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

func extract(dst string, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	switch filepath.Ext(src) {
	case ".zip":
		return extractZip(dst, f)
	case ".tar.gz":
		return extractTarGz(dst, f)
	default:
		return fmt.Errorf("unknown file extension: %v", filepath.Ext(src))
	}
}

func RemoveOldExtractions(prevCfg, newCfg *Config) error {
	for _, pa := range prevCfg.Applications {
		found := false
		if !found {
			log.Println("[extractions] remove", pa.ApplicationPath())
			err := os.RemoveAll(pa.ApplicationPath())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func AddNewExtractions(prevCfg, newCfg *Config) error {
	for _, na := range newCfg.Applications {
		found := false
		if !found {
			log.Println("[extractions] add", na.ApplicationPath())
			err := extract(na.ApplicationPath(), na.DownloadPath())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
