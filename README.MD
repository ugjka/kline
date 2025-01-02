# /////////////////////////////////////////////////////////////////

## MAIN.GO

```go
// ,--. ,--.,--.,--.                    ,---.   ,--.  ,---.   ,--.
// |  .'   /|  |`--',--,--,  ,---.     '.-.  \ /    \'.-.  \ /  .'
// |  .   ' |  |,--.|      \| .-. :     .-' .'|  ()  |.-' .'|  .-.
// |  |\   \|  ||  ||  ||  |\   --.    /   '-. \    //   '-.\   o |
// `--' '--'`--'`--'`--”--' `----'    '-----'  `--' '-----' `---'
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

```

# /////////////////////////////////////////////////////////////////

## LICENSE FOR THE CODE

Copyright © 2025 UGIS GERMANIS

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

## LICENSE FOR THE ART FOLDER

INCLUDED AS AN EXAMPLE/STARTER PACK, IS POSSIBLY NOT LICENSED, POSSIBLE COPYRIGHTED, BEWARE: YOU MAY GET SUED BIG TIME

## BUILDING

ASK CHATGPT HOW TO RUN GO CODE (HINT: YOU NEED THE LATEST GO VERSION)

## CREDITS

MUCH THANKS TO LIBERA HACKERS WHO EXPLAINED HOW THIS WORKS