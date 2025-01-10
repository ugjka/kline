package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"golang.org/x/text/encoding/ianaindex"
)

func main() {
	// pass filename as arg
	file := os.Args[1]
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	m := &matrix{}
	enc, err := ianaindex.IANA.Encoding("IBM437")
	if enc != nil && err == nil {
		data, err = enc.NewDecoder().Bytes(data)
		if err != nil {
			log.Fatal(err)
		}
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		var ansion bool
		text := []rune(string(data))
		var codes string
		for i := 0; i < len(text); i++ {
			// ansi prefix
			if i+1 < len(text) && text[i] == esc && text[i+1] == '[' {
				i++
				ansion = true
				continue
			}
			// ansi suffix
			if ansion && text[i] == 'm' {
				ansion = false
				parse(m, codes)
				codes = ""
				continue
			}
			// gather ansi codes
			if ansion {
				codes += string(text[i])
				continue
			}
			// strip dumbass control chars
			if text[i] == 10 || text[i] > 31 {
				m.addrune(text[i])
			} else {
				m.addrune(' ')
			}
		}
	}
	m.toirc()
	var ansicodes []int
	for k := range unhandled {
		ansicodes = append(ansicodes, k)
	}
	sort.Ints(ansicodes)
	if len(ansicodes) > 0 {
		fmt.Fprintln(os.Stderr, "unhandled ansi codes:", ansicodes)
	}
}

var unhandled = make(map[int]struct{})

func parse(m *matrix, codes string) {
	var nums []int
	for _, str := range strings.Split(codes, ";") {
		var i int
		_, err := fmt.Sscan(str, &i)
		if err != nil {
			log.Fatal(err)
		}
		nums = append(nums, i)
	}
	sort.Ints(nums)
	for _, num := range nums {
		switch {
		case num == 0:
			m.reset()
		case num == 1:
			m.boldon()
		case num >= 30 && num <= 37:
			m.fgset(num - 30)
		case num >= 40 && num <= 47:
			m.bgset(num - 40)
		default:
			unhandled[num] = struct{}{}
		}
	}
}

type cell struct {
	char rune
	bold bool
	bg   int
	fg   int
}

type matrix struct {
	cells   [][]cell
	nowbg   int
	nowfg   int
	nowbold bool
	cols    int
	row     int
}

func (m *matrix) bareprint() {
	for i := range m.cells {
		for j := range m.cells[i] {
			fmt.Printf("%c", m.cells[i][j].char)
		}
		fmt.Println()
	}
}

func (m *matrix) toirc() {
	for _, row := range m.cells {
		var bold bool
		var fg int
		var bg int
		for i, cell := range row {
			if i == 0 {
				// todo: skip background if the same as previous
				bold = cell.bold
				if cell.bold {
					fmt.Print("\x02")
					fg = bold2irc[cell.fg]
				} else {
					fg = ans2irc[cell.fg]
				}
				bg = ans2irc[cell.bg]
				//todo: only the bg needs to be 2 digits
				fmt.Printf("\x03%02d,%02d", fg, bg)
				fmt.Printf("%c", cell.char)
				continue
			}
			if bold != cell.bold {
				bold = cell.bold
				fmt.Print("\x02")
			}
			if bold {
				if fg != bold2irc[cell.fg] || bg != ans2irc[cell.bg] {
					fg = bold2irc[cell.fg]
					bg = ans2irc[cell.bg]
					fmt.Printf("\x03%02d,%02d", fg, bg)
				}
			} else {
				if fg != ans2irc[cell.fg] || bg != ans2irc[cell.bg] {
					fg = ans2irc[cell.fg]
					bg = ans2irc[cell.bg]
					fmt.Printf("\x03%02d,%02d", fg, bg)
				}
			}
			fmt.Printf("%c", cell.char)
			if i == len(row)-1 && len(row) < 80 {
				spaces := strings.Repeat(" ", 80-len(row))
				fmt.Printf("\x03%02d,%02d%s", ans2irc[7], ans2irc[0], spaces)
			}
		}
		fmt.Println()
	}
}

func (m *matrix) addrune(r rune) {
	if m.cells == nil {
		m.cells = make([][]cell, 0)
		m.cells = append(m.cells, make([]cell, 0))
	}
	if r == '\n' {
		m.cells = append(m.cells, make([]cell, 0))
		m.row++
		return
	}
	c := cell{
		char: r,
		bold: m.nowbold,
		bg:   m.nowbg,
		fg:   m.nowfg,
	}
	if len(m.cells[m.row]) == 80 {
		m.cells = append(m.cells, make([]cell, 0))
		m.row++
	}
	m.cells[m.row] = append(m.cells[m.row], c)
}

func (m *matrix) boldon() {
	m.nowbold = true
}

func (m *matrix) bgset(i int) {
	m.nowbg = i
}

func (m *matrix) fgset(i int) {
	m.nowfg = i
}

func (m *matrix) reset() {
	m.nowbold = false
	m.nowfg = 7
	m.nowbg = 0
}

const cols = 80
const esc rune = '\x1b'

var ans2irc = []int{
	88,
	40,
	44,
	41,
	48,
	50,
	46,
	96,
}

var bold2irc = []int{
	94,
	64,
	56,
	54,
	72,
	74,
	70,
	00,
}
