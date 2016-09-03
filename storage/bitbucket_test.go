package storage

import (
	"reflect"
	"testing"
)

func TestBitBucket(t *testing.T) {
	cases := []struct {
		input  string
		expect Location
	}{
		{
			"bitbucket://owner/repo/some/file.txt",
			Location{
				"scheme": "https",
				"host":   "api.bitbucket.org",
				"path":   "/1.0/repositories/owner/repo/raw/master/some/file.txt",
				"query":  "",
				"type":   "bitbucket",
			},
		},
		{
			"bitbucket://owner/repo/some/file.txt?branch=test",
			Location{
				"scheme": "https",
				"host":   "api.bitbucket.org",
				"path":   "/1.0/repositories/owner/repo/raw/test/some/file.txt",
				"query":  "",
				"type":   "bitbucket",
			},
		},
	}

	for _, tc := range cases {
		loc, _ := ParseLocation(tc.input)
		output := BitBucket.Build(loc)
		if !reflect.DeepEqual(output, tc.expect) {
			t.Errorf("for: %v\nexpected: %v\n     got: %v",
				tc.input, tc.expect, output)
		}
	}
}
