// ██╗  ██╗██╗     ██╗███╗   ██╗███████╗    ██████╗  ██████╗ ██████╗  ██████╗
// ██║ ██╔╝██║     ██║████╗  ██║██╔════╝    ╚════██╗██╔═████╗╚════██╗██╔════╝
// █████╔╝ ██║     ██║██╔██╗ ██║█████╗       █████╔╝██║██╔██║ █████╔╝███████╗
// ██╔═██╗ ██║     ██║██║╚██╗██║██╔══╝      ██╔═══╝ ████╔╝██║██╔═══╝ ██╔═══██╗
// ██║  ██╗███████╗██║██║ ╚████║███████╗    ███████╗╚██████╔╝███████╗╚██████╔╝
// ╚═╝  ╚═╝╚══════╝╚═╝╚═╝  ╚═══╝╚══════╝    ╚══════╝ ╚═════╝ ╚══════╝ ╚═════╝
package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/gogs/chardet"
	kitty "github.com/ugjka/kittybot"
	"golang.org/x/text/encoding/ianaindex"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Welcome to Kline, a botnet for spamming #libera-newyears
//
// kline commands as read from stdin:
// kline command "p art/asciiexample.txt" posts spam to partychan
// kline command "t art/asciiexample.txt" posts spam to testchan
// kline command "d milliseconds" sets  delay between messages
// kline command "a" aborts any current spamming
// kline command "l" lagtests the irc servers
//
// kline configuration below via constants

// bind kline to alternative address
// you probably don't want to run kline on your main ip address
// ipv6 for the win, also kline
//
// Examples:
// const BINDHOST = "2a01:4f7:c010:97e4::beec"
// const BINDHOST = "2a03:eb00:b9b3:11d3:2b83:9910:222f:1603"
const BINDHOST = ""

// set this to true when #libera-newyears is open for kliners
const PARTYCHANOPEN = false

// the official nye party chan on libera
// change at your own risk
const PARTYCHAN = "#libera-newyears"

// test channel
// make sure to voice the bots on the test chan
// otherwise you'll get K-Lined
const TESTCHAN = ""

// only spam in the test chan
// when all bots are voiced
// this guards against K-line
const TESTCHANGUARD = true

// kline nicks will start with this and end with some number
// don't forget to set this, ok?
const BOTNICKNAMEBASE = ""

// default kline message interval in ms
const DEFAULTDELAY = 100

// for shorter kline irc prefixes
// run fake ident server
// port 113 needs to be open and
// kline needs to be run as root
// or set bind cap on the binary as below:
// "sudo setcap 'cap_net_bind_service=ep' kline"
// turns off when all nicks served
const FAKEIDENTSERVER = false

// double the kline
const DOUBLEKLINE = false

func main() {
	//------------------------//
	// sewvews:
	// thewe awe mowe of these
	// these awe eu s-specific
	// dunnyo about the *boops your nose* usa fwiends
	// if you detect :3 some lag
	// you can try commenting them out
	// awso you can go on wibewa
	// and whois wandom peopwe to find mowe
	servers := []string{
		"zirconium.libera.chat", // Milano, IT
		"lead.libera.chat",      // Budapest, HU
		"tungsten.libera.chat",  // Umea, SE
		"platinum.libera.chat",  // Stockholm, SE
		"iridium.libera.chat",   // Stockholm, SE
		"erbium.libera.chat",    // Frankfurt, DE
		"osmium.libera.chat",    // Umea, SE
		"zinc.libera.chat",      // Espoo, FI
		"mercury.libera.chat",   // London, UK
		//"lithium.libera.chat", // Your, Shrink
	}

	if (!PARTYCHANOPEN || PARTYCHAN == "") && TESTCHAN == "" {
		fmt.Fprintf(os.Stderr, "error: no channels to work on")
		os.Exit(1)
	}

	if DOUBLEKLINE {
		// double the kline!
		servers = append(servers, servers...)
	}

	// launch fake ident server on kline
	if FAKEIDENTSERVER {
		go func() {
			count := len(servers)
			err := fakeIdentServer(BINDHOST, count)
			if err != nil {
				log.Error("kline", "fake ident error", err)
			}
		}()
	}

	// if BINDHOST set, creates a custom dialer for irc
	type dialfunc func(network string, addr string, tlsConf *tls.Config) (*tls.Conn, error)
	customTLSDial, err := func() (dialfunc, error) {
		if BINDHOST == "" {
			return nil, nil
		}

		localAddr, err := net.ResolveIPAddr("ip", BINDHOST)
		if err != nil {
			return nil, err
		}

		localTCPAddr := net.TCPAddr{
			IP: localAddr.IP,
		}

		dialer := &net.Dialer{
			LocalAddr: &localTCPAddr,
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}

		tlsdialer := func(network string, addr string, tlsConf *tls.Config) (*tls.Conn, error) {
			return tls.DialWithDialer(dialer, network, addr, &tls.Config{})
		}
		return tlsdialer, nil
	}()
	if err != nil {
		fmt.Fprintln(os.Stderr, "BINDHOST:", err)
		os.Exit(1)
	}

	var printdb chans
	var countVoicing atomic.Int64
	var testchanVoiced atomic.Bool
	testchanVoiced.Store(!TESTCHANGUARD)
	var bots []*kitty.Bot
	var once sync.Once
	var pingpong = make([]chan time.Time, len(servers))
	if DOUBLEKLINE {
		pingpong = make([]chan time.Time, len(servers)/2)
	}
	for i := range len(servers) {
		opts := func(bot *kitty.Bot) {
			bot.Channels = []string{}
			if PARTYCHAN != "" && PARTYCHANOPEN {
				bot.Channels = append(bot.Channels, PARTYCHAN)
			}
			if TESTCHAN != "" {
				bot.Channels = append(bot.Channels, TESTCHAN)
			}

			bot.SSL = true
			bot.ThrottleDelay = 0
			bot.DialTLS = customTLSDial
		}
		if BOTNICKNAMEBASE == "" {
			panic("you did something stupid, you silly willy")
		}
		nick := fmt.Sprintf("%s%02d", BOTNICKNAMEBASE, i)
		bot := kitty.NewBot(fmt.Sprintf("%s:6697", servers[i]), nick, opts)
		bot.SSL = true
		bot.Logger.SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StdoutHandler))

		if TESTCHANGUARD {
			bot.AddTrigger(kitty.Trigger{
				Condition: func(b *kitty.Bot, m *kitty.Message) bool {
					return m.Command == "MODE" && m.Param(0) == TESTCHAN && strings.HasPrefix(m.Param(1), "+v") && slices.Contains(m.Params[2:], nick)
				},
				Action: func(b *kitty.Bot, m *kitty.Message) {
					log.Info("kline", nick, "voiced in the test chan!", "remaining", len(servers)-int(countVoicing.Add(1)))
					if int(countVoicing.Load()) == len(servers) {
						log.Info("kline", "success", "all bots in the test channel voiced! You can proceed!")
						testchanVoiced.Store(true)
					}
				}})
		}

		// kline shut off valve for ops
		// on mode +m on the chan or bot
		// listened by first kline bot
		once.Do(func() {
			if PARTYCHAN == "" || !PARTYCHANOPEN {
				return
			}
			bot.AddTrigger(kitty.Trigger{
				Condition: func(b *kitty.Bot, m *kitty.Message) bool {
					return m.Command == "MODE" && m.Param(0) == PARTYCHAN && m.Param(1) == "+m" && (m.Param(3) == "" || m.Param(3) == nick)
				},
				Action: func(b *kitty.Bot, m *kitty.Message) {
					b.Logger.Warn("kline", "shut off valve", "engaged")
					printdb.clear(PARTYCHAN)
				},
			})
		})

		bot.AddTrigger(kitty.Trigger{
			Condition: func(b *kitty.Bot, m *kitty.Message) bool {
				return m.Command == "PONG"
			},
			Action: func(b *kitty.Bot, m *kitty.Message) {
				if i > len(pingpong) {
					return
				}
				select {
				case t, ok := <-pingpong[i]:
					if ok {
						log.Info("lag test", servers[i], time.Since(t))
					}
				default:
				}
			},
		})

		bots = append(bots, bot)
	}

	for _, b := range bots {
		go b.Run()
	}

	var delay atomic.Int64
	delay.Store(int64(time.Millisecond) * DEFAULTDELAY)

	// kline lazor printer made by brother
	go func() {
		i := 0
		var max int
		var tmp []byte
		for {
			time.Sleep(time.Duration(delay.Load()))
			printdb.Lock()
			for channel, line := range printdb.store {
				tmp = line.get()
				if tmp != nil {
					max = bots[i].MsgMaxSize(channel)
					if len(tmp) > max {
						tmp = tmp[:max]
					}
					bots[i].Msg(channel, string(tmp)) // printer goes brrrrrr
				}
			}
			printdb.Unlock()
			i++
			if i == len(bots) {
				i = 0
			}
		}
	}()

	// delicious kline
	spam := func(channel, file string) { // in a can
		// yummy!!!
		if strings.HasPrefix(file, "/dev/") {
			fmt.Fprintln(os.Stderr, "you a kline daredevil! But not today...")
			return
		}
		text, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		// attempts to
		// detect charset encoding and convert anything weird to utf-8
		// not exact science, may get garbage if something is niche
		if !utf8.Valid(text) {
			res, err := chardet.NewTextDetector().DetectBest(text)
			if err == nil {
				enc, err := ianaindex.IANA.Encoding(res.Charset)
				if enc != nil && err == nil {
					tmp, err := enc.NewDecoder().Bytes(text)
					if err == nil {
						text = tmp
					}
				}
			}

		}
		// schmucky object oriented programming
		lines := printdb.get(channel)
		lines.put(text)
	}

	// scan for kline bot commands from standard input, did you get your unix education yet?
	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		command := stdin.Text()
		parameters := strings.Split(command, " ")

		switch parameters[0] {
		// kline command "p <filename.txt>" posts spam to partychan
		case "p":
			if !PARTYCHANOPEN || PARTYCHAN == "" {
				fmt.Fprintln(os.Stderr, "error: partychan closed or not set")
				continue
			}
			if len(parameters) < 2 {
				fmt.Fprintln(os.Stderr, "error: missing file name")
				continue
			}
			spam(PARTYCHAN, strings.Join(parameters[1:], " "))

		// kline command "t <filename.txt>" posts spam to testchan
		case "t":
			if TESTCHAN == "" {
				fmt.Fprintln(os.Stderr, "error: testchan not set")
				continue
			}
			if !testchanVoiced.Load() {
				fmt.Fprintln(os.Stderr, "error: testchan bots are not voiced, will not proceed")
				continue
			}
			if len(parameters) < 2 {
				fmt.Fprintln(os.Stderr, "error: missing file name")
				continue
			}
			spam(TESTCHAN, strings.Join(parameters[1:], " "))

		// kline command "d <milliseconds>" sets  delay between messages
		case "d":
			if len(parameters) < 2 {
				fmt.Fprintln(os.Stderr, "error: missing delay")
				continue
			}
			var number int
			_, err := fmt.Sscanf(parameters[1], "%d", &number)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error setting delay:", err)
				continue
			}
			if time.Duration(number)*time.Millisecond > time.Second {
				fmt.Fprintln(os.Stderr, "error: delay can't be more than 1000ms")
				continue
			}
			delay.Store(int64(time.Millisecond) * int64(number))
			fmt.Fprintln(os.Stderr, "delay set to:", time.Duration(delay.Load()))

		// kline command "a" aborts current spamming
		case "a":
			printdb.clear("")

		// kline command "l" checks irc server lag time
		case "l":
			for i := range pingpong {
				pingpong[i] = make(chan time.Time, 1)
				pingpong[i] <- time.Now()
				bots[i].Send(fmt.Sprintf("PING :%s", servers[i]))
			}
		default:
			fmt.Println("error: invalid command")
		}
	}
}

// this was simpler than i imagined
// perhaps i should get a job!
func fakeIdentServer(bindaddress string, count int) error {
	localAddr, err := net.ResolveIPAddr("ip", bindaddress)
	if err != nil {
		return err
	}

	localTCPAddr := net.TCPAddr{
		IP:   localAddr.IP,
		Port: 113,
	}

	listener, err := net.ListenTCP("tcp", &localTCPAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Info("kline", "fake ident server running on", localAddr)

	var unixuser = 'a'
	var mu sync.Mutex
	var userwg sync.WaitGroup
	userwg.Add(count)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go func() {
			defer conn.Close()
			defer userwg.Done()
			r := bufio.NewReader(conn)
			request, err := r.ReadString('\n')
			if err != nil {
				return
			}

			var id1 int
			var id2 int
			_, err = fmt.Sscanf(request, "%d , %d\r\n", &id1, &id2)
			if err != nil {
				return
			}

			mu.Lock()
			response := fmt.Sprintf("%d, %d : USERID : UNIX : %c\r\n", id1, id2, unixuser)
			unixuser++
			if unixuser > 'z' {
				unixuser = 'a'
			}
			mu.Unlock()
			log.Info("kline", "got ident request", request, "response", response)
			conn.Write([]byte(response))
		}()
		count--
		if count == 0 {
			break
		}
	}
	userwg.Wait()
	log.Info("kline", "fake ident finished", "shutting down")
	return nil
}

// lines of coke
type lines struct {
	sync.Mutex
	lines    [][]byte
	tmplines [][]byte
	tmpline  []byte
}

// sniff it in
func (l *lines) put(data []byte) {
	l.Lock()
	l.tmplines = bytes.Split(data, []byte("\n"))
	l.lines = append(l.lines, l.tmplines...)
	l.Unlock()
}

// get your demons out
func (l *lines) get() []byte {
	l.Lock()
	defer l.Unlock()
	if len(l.lines) > 0 {
		l.tmpline = l.lines[0]
		l.lines = l.lines[1:]
		return l.tmpline
	} else {
		return nil
	}
}

// no fortune for you
type chans struct {
	sync.Mutex
	store map[string]*lines
}

// dibs
func (c *chans) get(ch string) *lines {
	c.Lock()
	defer c.Unlock()
	if c.store == nil {
		c.store = make(map[string]*lines)
	} else if l, ok := c.store[ch]; ok {
		return l
	}

	l := &lines{}
	c.store[ch] = l
	return l
}

// no dibs
func (c *chans) clear(ch string) {
	c.Lock()
	defer c.Unlock()
	if ch != "" {
		delete(c.store, ch)
		return
	}
	clear(c.store)
}
