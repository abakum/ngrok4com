package main

import (
	"embed"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xlab/closer"
	"go.bug.st/serial/enumerator"
)

const (
	EMULATOR = "com0com - serial port emulator"
	BIN      = "bin"
	BAUD     = "75"
	// BAUD = "19200"
	// BAUD  = "921600"
	LIMIT = "1"
	ITO   = "10"
	XO    = "on"
)

var (
	//go:embed NGROK_AUTHTOKEN.txt
	NGROK_AUTHTOKEN string
	//go:embed NGROK_API_KEY.txt
	NGROK_API_KEY string
	//go:embed bin/*
	bin embed.FS

	crypt,
	cwd,
	serial,
	ifs,
	publicURL,
	forwardsTo string
	port    = "7000"
	hub4com = `hub4com.exe`
	kitty   = `kitty_portable.exe`
	err,
	errNgrok error
	opts  = []string{"--baud=" + BAUD}
	ports []*enumerator.PortDetails
	ips   []string
	TO    = time.Second * 60
)

func main() {
	defer closer.Close()

	closer.Bind(func() {
		if err != nil {
			let.Println(err)
			defer os.Exit(1)
		}
		for _, v := range []string{
			"Proxies",
			"Sessions",
			// "kitty.ini",
		} {
			os.RemoveAll(filepath.Join(cwd, BIN, v))
		}
		// pressEnter()
	})

	NGROK_AUTHTOKEN = Getenv("NGROK_AUTHTOKEN", NGROK_AUTHTOKEN) //if emty then local mode
	// NGROK_AUTHTOKEN += "-"                                       // emulate bad token or no internet
	// NGROK_AUTHTOKEN = ""                                   // emulate local mode
	NGROK_API_KEY = Getenv("NGROK_API_KEY", NGROK_API_KEY) //if emty then no crypt
	// NGROK_API_KEY = ""                                     // emulate no crypt

	if NGROK_API_KEY != "" {
		crypt = "--create-filter=crypt,tcp,crypt:--secret=" + NGROK_API_KEY
	}

	ips = interfaces()
	if len(ips) == 0 {
		err = srcError(fmt.Errorf("not connected - нет сети"))
		return
	}
	ifs = strings.Join(ips, ",")

	cwd, err = os.Getwd()
	if err != nil {
		err = srcError(err)
		return
	}

	hub4com, err = write(hub4com)
	if err != nil {
		err = srcError(err)
		return
	}

	publicURL, forwardsTo, errNgrok = ngrokAPI(NGROK_API_KEY)
	ltf.Println(publicURL, forwardsTo, errNgrok)
	if len(os.Args) > 1 {
		i, er := strconv.Atoi(abs(os.Args[1]))
		if er != nil || i >= 9600 {
			// tty client mode
			// ngrok mode with encryption:
			// `ngrok4com baud` as `ngrok4com baud publicURL`
			// LAN mode with encryption:
			// `ngrok4com +` as `ngrok4com menuBaud 127.0.0.1`
			// `ngrok4com host` as `ngrok4com menuBaud host`
			// LAN mode without encryption:
			// `ngrok4com -baud` as `ngrok4com baud 127.0.0.1`
			// `ngrok4com -` as `ngrok4com menuBaud 127.0.0.1`
			// `ngrok4com -host` as `ngrok4com menuBaud host`
			tty()
			return
		}
		// serial server mode
		// case port omit it is 7000
		// ngrok mode with encryption:
		// `ngrok4com 0` as `ngrok4com menuSerial 7000`
		// `ngrok4com serial` as `ngrok4com serial 7000`
		// LAN mode without encryption:
		// `ngrok4com -0` as `ngrok4com menuSerial 7000`
		// `ngrok4com -serial` as `ngrok4com serial 7000`
		com()
	}

	// ngrok4com
	if errNgrok == nil {
		// use ngrok as `ngrok4com menuBaud publicURL`
		tty()
	} else {
		// create ngrok with `ngrok4com menuSerial`
		com()
	}
}

func abs(s string) string {
	if strings.HasPrefix(s, "-") {
		NGROK_AUTHTOKEN = "" // no ngrok
		NGROK_API_KEY = ""   // no crypt
		crypt = ""
		return s[1:]
	}
	if strings.HasPrefix(s, "+") {
		return s[1:]
	}
	return s
}

func write(fn string) (binFN string, err error) {
	binFN = path.Join(BIN, fn)
	bytes, err := bin.ReadFile(binFN)
	if err != nil {
		err = srcError(err)
		return
	}
	binDir := filepath.Join(cwd, BIN)
	_, err = os.Stat(binDir)
	if err != nil {
		err = os.MkdirAll(binDir, 0666)
		if err != nil {
			err = srcError(err)
			return
		}
	}
	binFN = filepath.Join(binDir, fn)
	_, err = os.Stat(binFN)
	if err == nil {
		return
	}
	ltf.Println(binFN, len(bytes))
	err = os.WriteFile(binFN, bytes, 0666)
	return
}

func interfaces() (ifs []string) {
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, ifac := range ifaces {
			addrs, err := ifac.Addrs()
			if err != nil || ifac.Flags&net.FlagUp == 0 || ifac.Flags&net.FlagRunning == 0 || ifac.Flags&net.FlagLoopback != 0 {
				continue
			}
			for _, addr := range addrs {
				if strings.Contains(addr.String(), ":") {
					continue
				}
				ifs = append(ifs, addr.String())
			}
		}
	}
	return
}
