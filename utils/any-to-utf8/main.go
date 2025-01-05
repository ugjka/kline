package main

import (
	"fmt"
	"log"
	"os"

	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"
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
	enc, name := charset.Lookup(res.Charset)
	fmt.Fprintln(os.Stderr, enc, name)
	data, err = enc.NewDecoder().Bytes(data)
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(data)
}
