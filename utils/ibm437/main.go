package main

import (
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
			m.addrune(text[i])
		}
	}
	m.serialize()
}

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
			m.fgset(ans2mircmap[num])
		case num >= 40 && num <= 47:
			m.bgset(ans2mircmap[num])
		default:
			fmt.Fprintln(os.Stderr, "unhandled ansi:", num)
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

func (m *matrix) serialize() {

	for _, row := range m.cells {
		var bold bool
		var fg int
		var bg int
		for i, cell := range row {
			if i == 0 {
				bold = cell.bold
				if cell.bold {
					fmt.Print("\x02")
				}
				fg = cell.fg
				bg = cell.bg
				fmt.Printf("\x03%02d,%02d", cell.fg, cell.bg)
				fmt.Printf("%c", cell.char)
			}
			if bold != cell.bold {
				bold = cell.bold
				fmt.Print("\x02")
			}
			if fg != cell.fg || bg != cell.bg {
				fg = cell.fg
				bg = cell.bg
				fmt.Printf("\x03%02d,%02d", fg, bg)
			}
			fmt.Printf("%c", cell.char)
		}
		fmt.Println()
	}
}

func (m *matrix) addrune(r rune) {
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
	if m.cells == nil {
		m.cells = make([][]cell, 0)
		m.cells = append(m.cells, make([]cell, 0))
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
	m.nowfg = 15
	m.nowbg = 1
}

const cols = 80
const esc rune = '\x1b'

var ans2mircmap = map[int]int{
	30: 1,
	31: 4,
	32: 9,
	33: 8,
	34: 12,
	35: 13,
	36: 11,
	37: 15,
	40: 1,
	41: 5,
	42: 3,
	43: 7,
	44: 2,
	45: 6,
	46: 10,
	47: 14,
}
