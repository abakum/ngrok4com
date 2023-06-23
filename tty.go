package main

import (
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dixonwille/wmenu/v5"
	"github.com/xlab/closer"
	"go.bug.st/serial/enumerator"
)

func tty() {
	var (
		err       error
		hub4com   = `..\hub4com\hub4com.exe`
		kitty     = `..\kitty\kitty_portable.exe`
		sercfg    = "9600"
		host      = "10.161.115.189:" + port
		serial    = "11"
		CNCB      = "10"
		publicURL string
		tcp       *url.URL
	)
	defer closer.Close()

	closer.Bind(func() {
		if err != nil {
			let.Println(err)
			defer os.Exit(1)
		}
		pressEnter()
	})

	if len(os.Args) > 1 {
		sercfg = os.Args[1]
	} else {
		menu := wmenu.NewMenu("Choose baud")
		menu.Action(func(opts []wmenu.Opt) error {
			sercfg = opts[0].Text
			return nil
		})
		menu.Option("9600", 1, sercfg == "9600", nil)
		menu.Option("38400", 2, sercfg == "38400", nil)
		menu.Option("57600", 3, sercfg == "57600", nil)
		menu.Option("115200", 4, sercfg == "115200", nil)
		err = menu.Run()
		if err != nil {
			err = srcError(err)
			return
		}
	}

	if len(os.Args) > 2 {
		host = os.Args[2]
		if !strings.Contains(host, ":") {
			host += ":" + port
		}
	}

	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		err = srcError(err)
		return
	}
	pair := ""
	for _, sPort := range ports {
		li.Println(sPort.Name, sPort.Product)
		if strings.HasPrefix(sPort.Product, emulator) {
			if pair == "" {
				pair = string(sPort.Product[len(sPort.Product)-1])
			}
			if !strings.HasSuffix(sPort.Product, pair) {
				continue
			}
			XX := strings.TrimPrefix(sPort.Product, emulator+" CNC")
			if strings.HasPrefix(XX, "B") {
				CNCB = strings.TrimPrefix(sPort.Name, "COM")
			} else {
				serial = strings.TrimPrefix(sPort.Name, "COM")
			}
		}
	}
	if pair == "" {
		err = Errorf("not found %s", emulator)
		return
	}

	li.Println(os.Args[0], "sercfg host")
	li.Println(os.Args)
	li.Println("sercfg", sercfg)

	if len(os.Args) > 2 || NGROK_API_KEY == "" {
		li.Println("local mode")
	} else {
		li.Println("ngrok mode")
		publicURL, _, err = ngrokAPI()
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

	li.Println("host", host)
	li.Println("serial", serial)
	li.Println("CNCB", CNCB)

	cwd, err := os.Getwd()
	if err == nil {
		hub4com = filepath.Join(cwd, hub4com)
		kitty = filepath.Join(cwd, kitty)
	}

	hub := exec.Command(
		hub4com,
		"--baud=460800",

		"--create-filter=escparse,com,parse",
		"--create-filter=pinmap,com,pinmap:--rts=cts --dtr=dsr",
		"--create-filter=linectl,com,lc:--br=local --lc=local",
		"--add-filters=0:com",

		"--create-filter=telnet,tcp,telnet:--comport=client",
		"--create-filter=pinmap,tcp,pinmap:--rts=cts --dtr=dsr --break=break",
		"--create-filter=linectl,tcp,lc:--br=remote --lc=remote",
		crypt,
		"--add-filters=1:tcp",

		"--octs=off",
		`\\.\COM`+CNCB,

		"--use-driver=tcp",
		host,
	)
	hub.Stdout = os.Stdout
	hub.Stderr = os.Stderr
	closer.Bind(func() {
		if hub.Process != nil && hub.ProcessState == nil {
			PrintOk("hub4com Kill", hub.Process.Kill())
		}
	})
	go func() {
		err := hub.Run()
		if err != nil {
			PrintOk("hub4com Run", err)
			closer.Close()
		}
	}()

	ki := exec.Command(
		kitty,
		"-sercfg",
		sercfg,
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
