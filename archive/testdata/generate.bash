#!/bin/bash

main() {
  cat >testdata.gen.go <<EOF
package testdata

// Generated Test Data
const (
    Original = "hello world\n"
    Tar      = \`$(tar -c hello.txt | base64 --wrap=0)\`
    XZTar    = \`$(tar -J -c hello.txt | base64 --wrap=0)\`
    LZMATar  = \`$(tar --lzma -c hello.txt | base64 --wrap=0)\`
    ZIPTar   = \`$(tar -z -c hello.txt | base64 --wrap=0)\`
    BZIPTar  = \`$(tar -j -c hello.txt | base64 --wrap=0)\`
)

EOF
}


main
