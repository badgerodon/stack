package archive

import (
	"bufio"
	"io"

	"github.com/ulikunitz/xz"
	"github.com/ulikunitz/xz/lzma"
)

var (
	// LZMA extracts lzma archives
	LZMA lzmaExtractor
	// XZ extracts xz archives
	XZ xzExtractor
)

type (
	lzmaExtractor struct{}
	xzExtractor   struct{}
)

func (p lzmaExtractor) ExtractReader(dst string, src io.Reader) error {
	z, err := lzma.NewReader(bufio.NewReader(src))
	if err != nil {
		return err
	}
	return Tar.ExtractReader(dst, z)
}

func (p xzExtractor) ExtractReader(dst string, src io.Reader) error {
	z, err := xz.NewReader(bufio.NewReader(src))
	if err != nil {
		return err
	}
	return Tar.ExtractReader(dst, z)
}
