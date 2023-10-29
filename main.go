// git clone github.com/abakum/ngrok4com

// go get github.com/dixonwille/wmenu/v5
// go get github.com/xlab/closer
// go get github.com/f1bonacc1/glippy
// go install github.com/tc-hib/go-winres@latest
// go get go.bug.st/serial
// go get golang.ngrok.com/ngrok
// go get golang.org/x/sync
// go get gopkg.in/ini.v1
// go get github.com/cakturk/go-netstat/netstat

// go-winres init
// git tag v0.2.1-lw
// git push origin --tags

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
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
	DELAY = "0.05"
	TOS   = time.Second * 7
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
	commandDelay = DELAY
	port         = "7000"
	hub4com      = `hub4com.exe`
	processName  = hub4com
	kitty        = `kitty_portable.exe`
	kittyINI     = `kitty.ini`
	err,
	errNgrok error
	opts  = []string{"--baud=" + BAUD}
	ports []*enumerator.PortDetails
	ips   []string
	TO    = time.Second * 60
	hub,
	ki,
	ngr *exec.Cmd
	plus,
	ok bool
)

func main() {
	defer closer.Close()

	closer.Bind(func() {
		kill(ki)
		kill(ngr)
		kill(hub)
		if err != nil {
			let.Println(err)
			defer os.Exit(1)
			pressEnter()
		}
		// for _, v := range []string{
		// 	"Proxies",
		// 	"Sessions",
		// 	"Jumplist",
		// 	// "kitty.ini",
		// } {
		// 	os.RemoveAll(filepath.Join(cwd, BIN, v))
		// }
	})

	NGROK_AUTHTOKEN = Getenv("NGROK_AUTHTOKEN", NGROK_AUTHTOKEN) //if emty then local mode
	// NGROK_AUTHTOKEN += "-"                                       // emulate bad token or no internet
	// NGROK_AUTHTOKEN = ""                                   // emulate local mode
	NGROK_API_KEY = Getenv("NGROK_API_KEY", NGROK_API_KEY) //if emty then no crypt
	// NGROK_API_KEY = ""                                     // emulate no crypt

	if NGROK_API_KEY != "" {
		crypt = "--create-filter=crypt,tcp,crypt:--secret=" + NGROK_API_KEY
	}

	cwd, err = os.Getwd()
	if err != nil {
		err = srcError(err)
		return
	}

	err = fs.WalkDir(fs.FS(bin), BIN, func(unix string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		win := filepath.Join(append([]string{cwd}, strings.Split(unix, "/")...)...)
		if d.IsDir() {
			_, err = os.Stat(win)
			if err != nil {
				err = os.MkdirAll(win, 0666)
			}
			return err
		}
		bytes, err := bin.ReadFile(unix)
		if err != nil {
			return err
		}
		update := true
		switch path.Base(unix) {
		case hub4com:
			hub4com = win
		case kitty:
			kitty = win
		case kittyINI:
			kittyINI = win
			update = false
		}
		var size int64
		fi, err := os.Stat(win)
		if err == nil {
			size = fi.Size()
			if int64(len(bytes)) == size || !update {
				return nil
			}
		}
		ltf.Println(win, len(bytes), "->", size)
		return os.WriteFile(win, bytes, 0666)
	})
	if err != nil {
		return
	}

	ips = interfaces()
	if len(ips) == 0 {
		err = srcError(fmt.Errorf("not connected - нет сети"))
		return
	}
	ifs = strings.Join(ips, ",")

	publicURL, forwardsTo, errNgrok = ngrokAPI(NGROK_API_KEY)
	ltf.Println(publicURL, forwardsTo, errNgrok)
	if len(os.Args) > 1 {
		i, er := strconv.Atoi(abs(os.Args[1]))
		if er != nil || i >= 75 {
			// tty client mode

			// `ngrok4com +` as `ngrok4com menuBaud publicURL` ngrok mode + encryption
			// `ngrok4com baud` as `ngrok4com baud publicURL` ngrok mode + encryption
			// `ngrok4com -` as `ngrok4com menuBaud 127.0.0.1` loop mode - encryption
			// `ngrok4com -baud` as `ngrok4com baud 127.0.0.1` loop mode - encryption
			// `ngrok4com baud host` as `ngrok4com baud host` LAN mode + encryption
			// `ngrok4com -baud host` as `ngrok4com baud host` LAN mode - encryption
			// `ngrok4com host` as `ngrok4com menuBaud host` LAN mode + encryption
			// `ngrok4com -host` as `ngrok4com menuBaud host` LAN mode - encryption
			tty()
			return
		}
		// serial server mode

		// `ngrok4com 0` as `ngrok4com menuSerial 7000` ngrok mode + encryption
		// `ngrok4com serial` as `ngrok4com serial 7000` ngrok mode + encryption
		// `ngrok4com -0` as `ngrok4com menuSerial 7000` LAN mode - encryption
		// `ngrok4com -serial` as `ngrok4com serial 7000` LAN mode - encryption
		com()
	}

	// ngrok4com
	if errNgrok == nil {
		// used ngrok as `ngrok4com menuBaud publicURL`
		tty()
	} else {
		// created ngrok with `ngrok4com menuSerial`
		com()
	}
}

func abs(s string) string {
	minus := strings.HasPrefix(s, "-")
	plus = strings.HasPrefix(s, "+")
	if minus || plus {
		if !plus {
			NGROK_AUTHTOKEN = ""
			NGROK_API_KEY = ""
		}
		if minus {
			crypt = ""
		}
		return s[1:]
	}
	return s
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

func kill(c *exec.Cmd) {
	if c != nil && c.Process != nil && c.ProcessState == nil {
		PrintOk(cmd("Kill", c), c.Process.Kill())
	}
}
