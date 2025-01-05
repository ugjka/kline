package main

import (
	"fmt"
	"log"
	"os"
	"unicode/utf8"

	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"
)

func main() {
	path := "./"
	// pass dir as arg
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	dir, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range dir {
		if file.IsDir() {
			continue
		}
		data, err := os.ReadFile(path + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		if !utf8.Valid(data) {
			res, err := chardet.NewTextDetector().DetectBest(data)
			if err != nil {
				fmt.Fprintln(os.Stderr, err, "in:", file.Name())
				continue
			}
			enc, _ := charset.Lookup(res.Charset)
			fmt.Println(enc, "in:", file.Name())
		}
	}
}
