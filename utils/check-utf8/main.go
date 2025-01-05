package main

import (
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/gogs/chardet"
	"golang.org/x/text/encoding/ianaindex"
)

func main() {
	// pass file or files
	var files []string
	if len(os.Args) > 1 {
		files = os.Args[1:]
	} else {
		fmt.Fprintln(os.Stderr, "no files given!")
		os.Exit(1)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		if !utf8.Valid(data) {
			res, err := chardet.NewTextDetector().DetectBest(data)
			if err != nil {
				fmt.Fprintln(os.Stderr, err, "in:", file)
				continue
			}
			enc, err := ianaindex.IANA.Encoding(res.Charset)
			if err != nil || enc == nil {
				fmt.Fprintln(os.Stderr, res.Charset, "not supported:", file)
				continue
			}
			fmt.Println(res.Confidence, enc, "in:", file)
		}
	}
}
