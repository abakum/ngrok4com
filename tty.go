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
		baud,
		host,
		CNCB string
		tcp *url.URL
	)

	li.Println("tty mode - режим терминала")
	li.Println(os.Args[0], "{[-]baud [host]|[-]host}")
	li.Println(os.Args)

	kitty, err = write(kitty)
	if err != nil {
		err = srcError(err)
		return
	}

	if len(os.Args) > 1 {
		_, err = strconv.Atoi(os.Args[1])
		if err != nil {
			// ngrok4com host
			// ngrok4com -host
			host = abs(os.Args[1])
		} else {
			// ngrok4com -
			// ngrok4com +
			// ngrok4com baud
			// ngrok4com -baud
			baud = abs(os.Args[1])
			if len(os.Args) > 2 {
				// ngrok4com baud host
				// ngrok4com -baud host
				host = abs(os.Args[2])
			}
		}
	}

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
		err = Errorf("not found %s\n`setupc'\n`install 0 PortName=COM#,RealPortName=COM11,EmuBR=yes,AddRTTO=10,AddRITO=10 -`\n", EMULATOR)
		return
	}
	li.Println("serial", serial)

	CNCB = `\\.\` + CNCB
	li.Println("CNCB", CNCB)

	if crypt == "" || errNgrok != nil || host != "" || plus {
		li.Println("LAN mode - режим локальной сети")
	} else {
		li.Println("ngrok mode - режим ngrok")
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
	hub = exec.Command(hub4com, append(opts,
		"--add-filters=1:tcp",

		// "--use-driver=serial",
		"--octs=off",
		"--ito="+ITO,
		"--ox="+XO,
		"--ix="+XO,
		"--write-limit="+LIMIT,
		CNCB,

		"--use-driver=tcp",
		host,
	)...)

	var bBuffer bytes.Buffer
	hub.Stdout = &bBuffer
	hub.Stderr = &bBuffer
	go func() {
		li.Println(cmd("Run", hub))
		err = srcError(hub.Run())
		PrintOk(cmd("Close", hub), err)
		if err != nil {
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

	for {
		if baud == "" {
			baud = "9600"
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
		}
		// li.Println("baud", baud)

		ki = exec.Command(
			kitty,
			"-sercfg",
			baud,
			"-serial",
			"COM"+serial,
		)

		li.Println(cmd("Run", ki))
		err = srcError(ki.Run())
		PrintOk(cmd("Close", ki), err)
		baud = ""
	}
	// closer.Hold()
}
