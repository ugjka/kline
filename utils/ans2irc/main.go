// Convert ANSI art to irc colors
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/text/encoding/ianaindex"
)

var COLUMNS *int

func main() {
	COLUMNS = flag.Int("cols", 80, "column count in ansi artwork")

	flag.Parse()

	// pass filename as arg
	if len(flag.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "error: no file name given")
		os.Exit(1)
	}
	file := flag.Args()[0]

	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	enc, err := ianaindex.IANA.Encoding("IBM437")
	if enc == nil || err != nil {
		fmt.Fprintln(os.Stderr, "error: IBM437 encoding not supported!", err)
		os.Exit(1)
	}

	data, err = enc.NewDecoder().Bytes(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	text := []rune(string(data))

	var isansi bool
	var params string

	var unknownformat []int
	var unknownoperation []rune

	m := &matrix{}
	m.init()
	// parse loop
loop:
	for i := 0; i < len(text); i++ {
		// ansi prefix
		if i+1 < len(text) && text[i] == '\x1b' && text[i+1] == '[' {
			i++
			isansi = true
			continue
		}

		switch {
		// formatting
		case isansi && text[i] == 'm':
			isansi = false
			u := formatting(m, params)
			unknownformat = append(unknownformat, u...)
			params = ""
			continue loop

		// char forward
		case isansi && text[i] == 'C':
			var moves int
			_, err := fmt.Sscanf(params, "%d", &moves)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ansi char move:", err)
			} else {
				m.cursormove(moves)
			}
			isansi = false
			params = ""
			continue loop
		// move up
		case isansi && text[i] == 'A':
			var moves int
			_, err := fmt.Sscanf(params, "%d", &moves)
			if err != nil {
				m.cursorup(1)
			} else {
				m.cursorup(moves)
			}
			isansi = false
			params = ""
			continue loop
		// no op
		case isansi && (text[i] >= 'a' && text[i] <= 'z' || text[i] >= 'A' && text[i] <= 'Z'):
			unknownoperation = append(unknownoperation, text[i])
			isansi = false
			params = ""
			continue loop
		}

		// gather parameters
		if isansi {
			params += string(text[i])
			continue
		}

		if text[i] < 4 || text[i] == '\x1A' {
			m.addrune(' ')
		} else {
			m.addrune(text[i])
		}
	}

	m.format2irc()

	if len(unknownformat) > 0 {
		sort.Ints(unknownformat)
		fmt.Fprintln(os.Stderr, "unhandled ansi formatting:", lo.Uniq(unknownformat))
	}
	if len(unknownoperation) > 0 {
		sort.Ints([]int(unknownformat))
		fmt.Fprintf(os.Stderr, "unhandled ansi operation: %c\n", lo.Uniq(unknownoperation))
	}
}

func formatting(m *matrix, codes string) (unknown []int) {
	var nums []int
	for _, str := range strings.Split(codes, ";") {
		var i int
		_, err := fmt.Sscan(str, &i)
		if err != nil {
			fmt.Fprintln(os.Stderr, "formatting parser:", err)
		}
		nums = append(nums, i)
	}
	sort.Ints(nums)
	for _, num := range nums {
		switch {
		case num == 0:
			m.reset()
		case num == 1:
			m.setbold()
		case num >= 30 && num <= 37:
			m.setfg(num - 30)
		case num >= 40 && num <= 47:
			m.setbg(num - 40)
		default:
			unknown = append(unknown, num)
		}
	}
	return unknown
}

func (m *matrix) format2irc() {
	var bold bool = false
	var fg int = ans2irc[7]
	var bg int = ans2irc[0]
	var oldbg int = ans2irc[0]
	for _, row := range m.rows {
		for i, cell := range row {
			// init first char because irc doesn't
			// carry over formating to next line
			if i == 0 {
				if !cell.set {
					cell.char = ' '
				}
				if bold != cell.bold {
					bold = cell.bold
					fmt.Print("\x02")
				}
				if bold && fg != ansbold2irc[cell.fg] {
					fg = ansbold2irc[cell.fg]
				} else if !bold && fg != ans2irc[cell.fg] {
					fg = ans2irc[cell.fg]
				}
				if bg != ans2irc[cell.bg] {
					bg = ans2irc[cell.bg]
					oldbg = bg
				}
				fmt.Printf("\x03%02d,%02d%c", fg, bg, cell.char)
				_ = oldbg
				continue
			}

			if !cell.set {
				cell.char = ' '
			}

			if bold != cell.bold {
				fmt.Print("\x02")
				bold = cell.bold
			}

			// it is hard to think about
			switch {
			case bold && fg != ansbold2irc[cell.fg]:
				fg = ansbold2irc[cell.fg]
				fmt.Printf("\x03%02d", fg)
			case !bold && fg != ans2irc[cell.fg]:
				fg = ans2irc[cell.fg]
				fmt.Printf("\x03%02d", fg)
			case bg != ans2irc[cell.bg] && !(fg == ans2irc[cell.bg] && cell.char == ' '):
				fmt.Printf("\x03%02d", fg)
			}

			switch {
			case fg == ans2irc[cell.bg] && cell.char == ' ':
				cell.char = '█'
			case bg != ans2irc[cell.bg]:
				bg = ans2irc[cell.bg]
				if bg != oldbg {
					fmt.Printf(",%02d", bg)
				}
				oldbg = bg
			}

			fmt.Printf("%c", cell.char)
		}
		fmt.Println()
	}
}

func (m *matrix) init() {
	m.rows = make([][]cell, 0)
	m.newrow()
	m.nowfg = 7
	m.nowbg = 0
}

func (m *matrix) newrow() {
	m.rows = append(m.rows, make([]cell, *COLUMNS))
}

func (m *matrix) cursormove(i int) {
	m.curcol += i
	for range m.curcol / *COLUMNS {
		m.newrow()
		m.currow++
	}
	m.curcol = m.curcol % *COLUMNS
}

func (m *matrix) cursorup(i int) {
	m.currow -= i
	if m.currow < 0 {
		m.currow = 0
	}
}

func (m *matrix) addrune(r rune) {
	if r == '\n' {
		if len(m.rows)-1 == m.currow {
			m.newrow()
		}
		m.currow++
		m.curcol = 0
		return
	}
	c := cell{
		char: r,
		bold: m.nowbold,
		fg:   m.nowfg,
		bg:   m.nowbg,
		set:  true,
	}
	m.rows[m.currow][m.curcol] = c
	m.curcol++
	if m.curcol == *COLUMNS {
		if len(m.rows)-1 == m.currow {
			m.newrow()
		}
		m.currow++
		m.curcol = 0
	}
}

func (m *matrix) setbold() {
	m.nowbold = true
}

func (m *matrix) setbg(i int) {
	m.nowbg = i
}

func (m *matrix) setfg(i int) {
	m.nowfg = i
}

func (m *matrix) reset() {
	m.nowbold = false
	m.nowfg = 7
	m.nowbg = 0
}

type cell struct {
	char rune
	bold bool
	bg   int
	fg   int
	set  bool
}

type matrix struct {
	rows    [][]cell
	nowbg   int
	nowfg   int
	nowbold bool
	currow  int
	curcol  int
}

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

var ansbold2irc = []int{
	94,
	64,
	56,
	54,
	72,
	74,
	70,
	00,
}
