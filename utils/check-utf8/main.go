package main

import (
	"fmt"
	"log"
	"os"
	"unicode/utf8"
)

func main() {
	path := os.Args[1]
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
			fmt.Println("Invalid UTF-8 in:", file.Name())
		}
	}
}
