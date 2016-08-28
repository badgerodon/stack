package archive

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	uuid "github.com/satori/go.uuid"
)

type (
	// An Extractor implements archive extraction
	Extractor interface {
		Extract(dst, src string) error
	}
	// A ReaderExtractor extracts readers
	ReaderExtractor interface {
		ExtractReader(dst string, src io.Reader) error
	}
	extractorDef struct {
		suffix    string
		extractor interface{}
	}
	extractorDefs []extractorDef
)

func (eds extractorDefs) Len() int {
	return len(eds)
}
func (eds extractorDefs) Swap(i, j int) {
	eds[i], eds[j] = eds[j], eds[i]
}
func (eds extractorDefs) Less(i, j int) bool {
	return len(eds[i].suffix) < len(eds[j].suffix)
}

var extractors []extractorDef

// Register registers a new archive extractor. This makes it easy to extend this library from outside
func Register(suffix string, extractor interface{}) {
	extractors = append(extractors, struct {
		suffix    string
		extractor interface{}
	}{suffix, extractor})
	sort.Sort(extractorDefs(extractors))
}

// Extract extracts the given archive to the given destination
func Extract(dst, src string) error {
	// TODO: make this more efficient
	for _, ed := range extractors {
		if strings.HasSuffix(src, ed.suffix) {
			if e, ok := ed.extractor.(Extractor); ok {
				return e.Extract(dst, src)
			} else if re, ok := ed.extractor.(ReaderExtractor); ok {
				// if we don't directly implement the Extract interface, we can use ExtractReader
				f, err := os.Open(src)
				if err != nil {
					return err
				}
				defer f.Close()
				return re.ExtractReader(dst, f)
			}
		}
	}
	return fmt.Errorf("unknown archive format: %s", filepath.Ext(src))
}

// ExtractReader extracts the given reader to the given destination
func ExtractReader(dst, srcName string, src io.Reader) error {
	for _, ed := range extractors {
		if strings.HasSuffix(srcName, ed.suffix) {
			if re, ok := ed.extractor.(ReaderExtractor); ok {
				// if we support the ReaderExtractor interface, just use that
				return re.ExtractReader(dst, src)
			} else if e, ok := ed.extractor.(Extractor); ok {
				// otherwise we can use the regular Extractor interface with a temporary file
				fname := filepath.Join(os.TempDir(), uuid.NewV4().String())
				defer os.RemoveAll(fname)

				f, err := os.Create(fname)
				if err != nil {
					return err
				}
				_, err = io.Copy(f, src)
				f.Close()
				if err != nil {
					return err
				}
				return e.Extract(dst, fname)
			}
		}
	}
	return fmt.Errorf("unknown archive format: %s", filepath.Ext(srcName))
}

func init() {
	Register(".tar", Tar)
	Register(".tar.bz2", BZip2)
	Register(".tar.gz", GZip)
	Register(".tar.lz", LZMA)
	Register(".tar.lzma", LZMA)
	Register(".tar.xz", XZ)
	Register(".tgz", GZip)
	Register(".tb2", BZip2)
	Register(".tbz", BZip2)
	Register(".tbz2", BZip2)
	Register(".tlz", LZMA)
	Register(".txz", XZ)
	Register(".zip", Zip)
}
