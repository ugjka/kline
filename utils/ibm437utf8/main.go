package main

import (
	"bytes"
	"log"
	"os"

	"golang.org/x/text/encoding/ianaindex"
)

func main() {
	// pass filename as arg
	file := os.Args[1]
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	enc, err := ianaindex.IANA.Encoding("IBM437")
	if enc != nil && err == nil {
		data, err = enc.NewDecoder().Bytes(data)
		if err != nil {
			log.Fatal(err)
		}
	}
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte{'\n'})
	data = bytes.Split(data, []byte{'\x1A'})[0]
	buf := bytes.NewBuffer(nil)
	var cp437 = []rune("\x00☺☻♥♦♣♠•◘○◙♂♀♪♬☼►◄↕‼¶§▬↨↑↓→←∟↔▲▼")
	for _, r := range []rune(string(data)) {
		// replace control chars with chars from ibm437 set
		if r < 32 && r != '\x1b' && r != '\n' {
			buf.WriteRune(cp437[r])
		} else {
			buf.WriteRune(r)
		}
	}
	buf.WriteTo(os.Stdout)
}
