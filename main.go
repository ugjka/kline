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
	"strings"
	"sync"
	"sync/atomic"
	"time"

	kitty "github.com/ugjka/kittybot"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Welcome to Kline, a botnet for spamming #libera-newyears
//
// kline commands as read from stdin:
// kline command "p art/asciiexample.txt" posts spam to partychan
// kline command "t art/asciiexample.txt" posts spam to testchan
// kline command "d milliseconds" sets  delay between messages
// kline command "a" aborts any current spamming
//
// kline configuration below via constants

// bind kline to alternative address
// you probably don't want to run kline on your main ip address
// ipv6 for the win, also kline
//
// Examples:
// const BINDHOST = "2a01:4f8:c010:97e4::beec"
// const BINDHOST = "2a03:eb00:b9a3:11d3:2b83:9910:222f:1603"
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
// port 113 needs to be open
// kline needs to be run as root
// or do some CAP/perm mumbo jumbo
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
		"tungsten.libera.chat",  // Stockholm, SE
		"platinum.libera.chat",  // Stockholm, SE
		"iridium.libera.chat",   // Stockholm, SE
		"osmium.libera.chat",    // Umea, SE
		"zinc.libera.chat",      // Espoo, FI
		"mercury.libera.chat",   // London, UK
		"lithium.libera.chat",   // Your, Shrink
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

	var countVoicing atomic.Int64
	var testchanVoiced atomic.Bool
	testchanVoiced.Store(!TESTCHANGUARD)
	var bots []*kitty.Bot
	var once sync.Once
	var abortSpam atomic.Bool
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
			if BINDHOST != "" {
				bot.DialTLS = customTLSDial
			}
		}
		if BOTNICKNAMEBASE == "" {
			panic("you did something stupid, you silly willy")
		}
		nick := fmt.Sprintf("%s%02d", BOTNICKNAMEBASE, i)
		bot := kitty.NewBot(fmt.Sprintf("%s:6697", servers[i]), nick, opts)
		bot.SSL = true
		bot.Logger.SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StdoutHandler))

		if TESTCHANGUARD {
			var mu sync.Mutex
			bot.AddTrigger(kitty.Trigger{
				Condition: func(b *kitty.Bot, m *kitty.Message) bool {
					return m.Command == "MODE" && m.Param(0) == TESTCHAN && m.Param(1) == "+v" && m.Param(2) == nick
				},
				Action: func(b *kitty.Bot, m *kitty.Message) {
					mu.Lock()
					countVoicing.Add(1)
					log.Info("kline", m.Param(2), "voiced in the test chan!", "remaining", len(servers)-int(countVoicing.Load()))
					if int(countVoicing.Load()) == len(servers) {
						log.Info("kline", "success", "all bots in the test channel voiced! You can proceed!")
						testchanVoiced.Store(true)
					}
					mu.Unlock()
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
					abortSpam.Store(true)
				},
			})
			bot.AddTrigger(kitty.Trigger{
				Condition: func(b *kitty.Bot, m *kitty.Message) bool {
					return m.Command == "MODE" && m.Param(0) == PARTYCHAN && m.Param(1) == "-m" && (m.Param(3) == "" || m.Param(3) == nick)
				},
				Action: func(b *kitty.Bot, m *kitty.Message) {
					b.Logger.Warn("kline", "shut off valve", "disengaged")
					abortSpam.Store(false)
				},
			})
		})

		bots = append(bots, bot)
	}

	for _, b := range bots {
		go b.Run()
	}

	var delay atomic.Int64
	delay.Store(int64(time.Millisecond) * DEFAULTDELAY)

	// delicious
	spam := func(channel, file string) {
		text, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		lines := bytes.Split(text, []byte("\n"))
		i := 0
		for _, line := range lines {
			if abortSpam.Load() {
				return
			}
			time.Sleep(time.Duration(delay.Load()))
			bots[i].Msg(channel, string(line))
			if i == len(bots)-1 {
				i = 0
			} else {
				i++
			}
		}
	}

	// scan for kline bot commands from standard input
	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		command := stdin.Text()
		parametrs := strings.Split(command, " ")

		switch parametrs[0] {
		// kline command "p <filename.txt>" posts spam to partychan
		case "p":
			if !PARTYCHANOPEN || PARTYCHAN == "" {
				fmt.Fprintln(os.Stderr, "error: partychan closed or not set")
				continue
			}
			if len(parametrs) < 2 {
				fmt.Fprintln(os.Stderr, "error: missing file name")
				continue
			}
			go spam(PARTYCHAN, strings.Join(parametrs[1:], " "))

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
			if len(parametrs) < 2 {
				fmt.Fprintln(os.Stderr, "error: missing file name")
				continue
			}
			go spam(TESTCHAN, strings.Join(parametrs[1:], " "))

		// kline command "d <milliseconds>" sets  delay between messages
		case "d":
			if len(parametrs) < 2 {
				fmt.Fprintln(os.Stderr, "error: missing delay")
				continue
			}
			var number int
			_, err := fmt.Sscanf(parametrs[1], "%d", &number)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error setting delay:", err)
				continue
			}
			if time.Duration(number)*time.Millisecond > time.Second {
				fmt.Fprintln(os.Stderr, "error: delay can't be bigger than 1000ms")
				continue
			}
			delay.Store(int64(time.Millisecond) * int64(number))
			fmt.Fprintln(os.Stderr, "delay set to:", time.Duration(delay.Load()))

		// kline command "a" aborts current spamming
		case "a":
			abortSpam.Store(true)
			time.Sleep(time.Second + time.Duration(delay.Load()))
			abortSpam.Store(false)
		default:
			fmt.Println("error: invalid command")
		}
	}
}

// this was simpler than i imagined
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
