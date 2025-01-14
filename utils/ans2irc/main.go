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
	*COLUMNS--

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

	// split off 16colo.rs metadata
	// and stuff
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.Split(data, []byte{'\x1A'})[0]
	data = bytes.TrimRight(data, "\n")

	text := []rune(string(data))

	var isansi bool
	var params string

	var unknownformat []int
	var unknownoperation []rune

	m := &matrix{}
	m.init()

	//var irccontrol = []rune{'\x02', '\x1d', '\x1f', '\x1e', '\x11', '\x03', '\x04', '\x16', '\x0f'}

	// we only use the beginning of this but whatever it can stay in its entirety
	var cp437 = []rune("\x00☺☻♥♦♣♠•◘○◙♂♀♪♬☼►◄↕‼¶§▬↨↑↓→←∟↔▲▼ !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~⌂ÇüéâäàåçêëèïîìÄÅÉæÆôöòûùÿÖÜ¢£¥₧ƒáíóúñÑªº¿⌐¬½¼¡«»░▒▓│┤╡╢╖╕╣║╗╝╜╛┐└┴┬├─┼╞╟╚╔╩╦╠═╬╧╨╤╥╙╘╒╓╫╪┘┌█▄▌▐▀αßΓπΣσµτΦΘΩδ∞φε∩≡±≥≤⌠⌡÷≈°∙·√ⁿ²■\u00A0")

	// the big parse loop
loop:
	for i := 0; i < len(text); i++ {
		// ansi prefix
		if i+1 < len(text) && text[i] == '\x1b' && text[i+1] == '[' {
			i++
			if isansi == true {
				// strip incomplete sequences
				defer fmt.Fprintf(os.Stderr, "stripped incomplete ansi params: %#x\n", params)
				params = ""
			}
			isansi = true
			continue
		}

		// finding test cases example:
		// find . -type f -exec grep -IqP "\x1b\[\d{2}B" {} \; -print
		// from
		// https://github.com/sixteencolors/sixteencolors-archive

		switch {
		// save cursor position
		case isansi && text[i] == 's':
			m.save()
			isansi = false
			params = ""
			continue loop
		// restore cursor position
		case isansi && text[i] == 'u':
			m.restore()
			isansi = false
			params = ""
			continue loop
		// set cursor position
		case isansi && text[i] == 'H':
			err := m.position(params)
			if err != nil {
				defer fmt.Fprintln(os.Stderr, "ansi H:", err)
			}
			isansi = false
			params = ""
			continue loop
		// cursor up
		case isansi && text[i] == 'A':
			var moves int
			if params == "" {
				m.up(1)
			} else if _, err := fmt.Sscanf(params, "%d", &moves); err != nil {
				defer fmt.Fprintln(os.Stderr, "ansi A:", err)
			} else {
				m.up(moves)
			}
			isansi = false
			params = ""
			continue loop
		// cursor down
		case isansi && text[i] == 'B':
			var moves int
			if params == "" {
				m.down(1)
			} else if _, err := fmt.Sscanf(params, "%d", &moves); err != nil {
				defer fmt.Fprintln(os.Stderr, "ansi B:", err)
			} else {
				m.down(moves)
			}
			isansi = false
			params = ""
			continue loop
		// cursor forward
		case isansi && text[i] == 'C':
			var moves int
			if params == "" {
				m.forward(1)
			} else if _, err := fmt.Sscanf(params, "%d", &moves); err != nil {
				defer fmt.Fprintln(os.Stderr, "ansi C:", err)
			} else {
				m.forward(moves)
			}
			isansi = false
			params = ""
			continue loop
		// cursor backward
		case isansi && text[i] == 'D':
			var moves int
			if params == "" {
				m.backward(1)
			} else if _, err := fmt.Sscanf(params, "%d", &moves); err != nil {
				defer fmt.Fprintln(os.Stderr, "ansi D:", err)
			} else {
				m.backward(moves)
			}
			isansi = false
			params = ""
			continue loop
		// formatting
		case isansi && text[i] == 'm':
			u, errs := formatting(m, params)
			defer func() {
				for _, err := range errs {
					fmt.Fprintln(os.Stderr, "formatting:", err)
				}
			}()
			unknownformat = append(unknownformat, u...)
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
			continue loop
		}

		// replace some escape codes with ibm437 set
		if text[i] < 32 && text[i] != '\x1b' && text[i] != '\n' {
			m.addrune(cp437[text[i]])
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

func formatting(m *matrix, codes string) (unknown []int, errs []error) {
	var nums []int
	for _, str := range strings.Split(codes, ";") {
		var i int
		_, err := fmt.Sscan(str, &i)
		if err != nil {
			errs = append(errs, err)
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
	return unknown, errs
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
	m.rows = append(m.rows, make([]cell, *COLUMNS+1))
}

func (m *matrix) forward(i int) {
	m.curcol += i
	if m.curcol > *COLUMNS {
		m.curcol = *COLUMNS
	}
}

func (m *matrix) backward(i int) {
	m.curcol -= i
	if m.curcol < 0 {
		m.curcol = 0
	}
}

func (m *matrix) up(i int) {
	m.currow -= i
	if m.currow < 0 {
		m.currow = 0
	}
}

func (m *matrix) down(i int) {
	m.currow += i
	i = m.currow - len(m.rows) + 1
	for range i {
		m.newrow()
	}
}

func (m *matrix) position(codes string) (err error) {
	if strings.HasPrefix(codes, ";") {
		codes = "1" + codes
	}
	if strings.HasSuffix(codes, ";") {
		codes = codes + "1"
	}
	if codes == "" {
		codes = "1;1"
	}
	var row int
	var col int
	_, err = fmt.Sscanf(codes, "%d;%d", &row, &col)
	row -= 1
	col -= 1
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}
	if row > len(m.rows)-1 {
		tmp := row - len(m.rows) + 1
		for range tmp {
			m.newrow()
		}
	}
	if col > *COLUMNS {
		col = *COLUMNS
	}
	m.currow = row
	m.curcol = col
	return err
}

// TODO: do not print extra empty line
// something something
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
	if m.curcol > *COLUMNS {
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

func (m *matrix) save() {
	m.tmpcol = m.curcol
	m.tmprow = m.currow
}

func (m *matrix) restore() {
	m.curcol = m.tmpcol
	m.currow = m.tmprow
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
	tmprow  int
	tmpcol  int
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
