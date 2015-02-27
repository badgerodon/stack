package archive

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type (
	ArchiveProvider interface {
		Extract(dst, src string) error
		Create(dst, src string) error
	}
	archiveProviderDef struct {
		suffix   string
		provider ArchiveProvider
	}
	archiveProviderDefs []archiveProviderDef
)

func (apd archiveProviderDefs) Len() int {
	return len(apd)
}
func (apd archiveProviderDefs) Swap(i, j int) {
	apd[i], apd[j] = apd[j], apd[i]
}
func (apd archiveProviderDefs) Less(i, j int) bool {
	return len(apd[i].suffix) < len(apd[j].suffix)
}

var archiveProviders = []archiveProviderDef{}

func Register(suffix string, provider ArchiveProvider) {
	archiveProviders = append(archiveProviders, struct {
		suffix   string
		provider ArchiveProvider
	}{suffix, provider})
	sort.Sort(archiveProviderDefs(archiveProviders))
}

func Create(dst, src string) error {
	for _, apd := range archiveProviders {
		if strings.HasSuffix(src, apd.suffix) {
			return apd.provider.Create(dst, src)
		}
	}
	return fmt.Errorf("unknown archive format: %s", filepath.Ext(src))
}

func Extract(dst, src string) error {
	for _, apd := range archiveProviders {
		if strings.HasSuffix(src, apd.suffix) {
			return apd.provider.Extract(dst, src)
		}
	}
	return fmt.Errorf("unknown archive format: %s", filepath.Ext(src))
}
