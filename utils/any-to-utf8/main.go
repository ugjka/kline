package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gogs/chardet"
	"golang.org/x/text/encoding/ianaindex"
)

func main() {
	// pass filename as arg
	file := os.Args[1]
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	res, err := chardet.NewTextDetector().DetectBest(data)
	if err != nil {
		log.Fatal(err)
	}
	enc, err := ianaindex.IANA.Encoding(res.Charset)
	if enc != nil && err == nil {
		fmt.Fprintln(os.Stderr, enc, res.Confidence, file)
		data, err = enc.NewDecoder().Bytes(data)
		if err != nil {
			log.Fatal(err)
		}
		os.Stdout.Write(data)
		return
	}
	fmt.Fprintln(os.Stderr, "encoding not detected")
}
