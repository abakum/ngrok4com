package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/dixonwille/wmenu/v5"
	"github.com/xlab/closer"
	"go.bug.st/serial/enumerator"
)

func tty() {
	var (
		baud = "9600"
		host = "127.0.0.1"
		CNCB,
		publicURL string
		tcp *url.URL
	)

	li.Println("tty mode - режим терминала")
	li.Println(os.Args[0], "[-][baud] [host]")
	li.Println(os.Args)

	kitty, err = write(kitty)
	if err != nil {
		err = srcError(err)
		return
	}

	if len(os.Args) > 1 {
		_, err = strconv.Atoi(os.Args[1])
		if err != nil {
			host = abs(os.Args[1])
			menu := wmenu.NewMenu("Choose baud - Выбери скорость")
			menu.Action(func(opts []wmenu.Opt) error {
				baud = opts[0].Text
				return nil
			})
			menu.Option("9600", 1, baud == "9600", nil)
			menu.Option("38400", 2, baud == "38400", nil)
			menu.Option("57600", 3, baud == "57600", nil)
			menu.Option("115200", 4, baud == "115200", nil)
			err = menu.Run()
			if err != nil {
				err = srcError(err)
				return
			}
		} else {
			baud = abs(os.Args[1])
			if len(os.Args) > 2 {
				host = os.Args[2]
			}
		}
	}

	li.Println("baud", baud)

	ports, err = enumerator.GetDetailedPortsList()
	if err != nil {
		err = srcError(err)
		return
	}
	pair := ""
	for _, sPort := range ports {
		// look only first pair
		com := strings.TrimPrefix(sPort.Name, "COM")
		if strings.HasPrefix(sPort.Product, EMULATOR) {
			li.Println(sPort.Name, sPort.Product)
			if strings.HasPrefix(sPort.Product, EMULATOR+" CNC") {
				// Windows10
				p := string(strings.TrimPrefix(sPort.Product, EMULATOR+" CNC")[1])
				if pair == "" {
					pair = p
				}
				if pair != p {
					continue
				}
				if strings.HasPrefix(sPort.Product, EMULATOR+" CNCA") {
					// setupc install PortName=sPort.Name -
					serial = com
					CNCB = "CNCB" + pair
				} else {
					// setupc install PortName=COMserial PortName=sPort.Name
					CNCB = sPort.Name
					break
				}
			} else {
				// Windows7
				if serial == "" {
					serial = com
					CNCB = "CNCB0"
				} else {
					CNCB = sPort.Name
					break
				}
			}
		}
	}
	if serial == "" {
		err = Errorf("not found %s\n`install 0 PortName=COM11,EmuBR=yes -`\n", EMULATOR)
		return
	}
	li.Println("serial", serial)

	CNCB = `\\.\` + CNCB
	li.Println("CNCB", CNCB)

	if host != "" || NGROK_API_KEY == "" {
		NGROK_AUTHTOKEN = "" // no ngrok
		NGROK_API_KEY = ""   // no crypt
		li.Println("LAN mode - режим локальной сети")
	} else {
		li.Println("ngrok mode - режим ngrok")
		publicURL, _, err = ngrokAPI(NGROK_API_KEY)
		if err != nil {
			return
		}

		tcp, err = url.Parse(publicURL)
		if err != nil {
			err = srcError(err)
			return
		}
		host = tcp.Host
	}
	if !strings.Contains(host, ":") {
		host += ":" + port
	}
	li.Println("host", host)

	opts = append(opts,
		"--create-filter=escparse,com,parse",
		"--create-filter=pinmap,com,pinmap:--rts=cts --dtr=dsr",
		"--create-filter=linectl,com,lc:--br=local --lc=local",
		"--add-filters=0:com",

		"--create-filter=telnet,tcp,telnet:--comport=client",
		"--create-filter=pinmap,tcp,pinmap:--rts=cts --dtr=dsr --break=break",
		"--create-filter=linectl,tcp,lc:--br=remote --lc=remote",
	)
	if crypt != "" {
		opts = append(opts, crypt)
	}
	hub := exec.Command(hub4com, append(opts,
		"--add-filters=1:tcp",

		"--octs=off",
		CNCB,

		"--use-driver=tcp",
		host,
	)...)

	var bBuffer bytes.Buffer
	hub.Stdout = &bBuffer
	hub.Stderr = &bBuffer
	closer.Bind(func() {
		if hub.Process != nil && hub.ProcessState == nil {
			PrintOk("hub4com Kill", hub.Process.Kill())
		}
	})
	go func() {
		err = hub.Run()
		if err != nil {
			PrintOk("hub4com Run", err)
			closer.Close()
		}
	}()
	for i := 0; i < 24; i++ {
		s, er := bBuffer.ReadString('\n')
		if er == nil {
			if strings.Contains(s, "ERROR") {
				err = Errorf(s)
				return
			}
			fmt.Print(s)
			if s == "TCP(1): Connected\n" {
				break
			}
		}
		time.Sleep(time.Millisecond * 50)
	}
	// fmt.Print(bBuffer.String())
	bBuffer.WriteTo(os.Stdout)
	hub.Stdout = os.Stdout
	hub.Stderr = os.Stderr

	ki := exec.Command(
		kitty,
		"-sercfg",
		baud,
		"-serial",
		"COM"+serial,
	)
	closer.Bind(func() {
		if ki.Process != nil && ki.ProcessState == nil {
			PrintOk("kitty Kill", ki.Process.Kill())
		}
	})

	PrintOk("kitty Run", ki.Run())
}
