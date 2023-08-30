package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dixonwille/wmenu/v5"
	"github.com/xlab/closer"
	"go.bug.st/serial/enumerator"
)

func com() {
	var (
		ngrokBin = `..\ngrok\ngrok.exe`
	)

	li.Println("serial server mode - режим сервера порта")
	li.Println(os.Args[0], "[-]serial [port]")
	li.Println(os.Args)

	ngrokBin = filepath.Join(cwd, ngrokBin)

	if len(os.Args) > 1 {
		serial = abs(os.Args[1])
	}

	if len(os.Args) > 2 {
		port = os.Args[2]
	}

	ports, err = enumerator.GetDetailedPortsList()
	if err != nil {
		err = srcError(err)
		return
	}
	i := 0
	isDefault := false
	ok := false
	menu := wmenu.NewMenu("Choose serial port- Выбери последовательный порт")
	for _, sPort := range ports {
		title := fmt.Sprintf("%s %s", sPort.Name, sPort.Product)
		if !strings.Contains(sPort.Product, EMULATOR) {
			li.Println(title)
			value := strings.TrimPrefix(sPort.Name, "COM")
			if serial == "" || serial == "0" {
				serial = value
			} else {
				isDefault = serial == value
			}
			ok = ok || isDefault
			if i == 1 && !ok {
				serial = value
				isDefault = true
			}
			menu.Option(title, value, isDefault, nil)
			i++
		}
	}
	switch i {
	case 0:
		err = Errorf("no serial port")
		return
	case 1:
	default:
		if !ok {
			menu.Action(func(opts []wmenu.Opt) error {
				serial = opts[0].Value.(string)
				return nil
			})
			err = menu.Run()
			if err != nil {
				err = srcError(err)
				return
			}
		}
	}

	li.Println("serial", serial)
	li.Println("port", port)

	opts = append(opts,
		// "--interface=127.0.0.1",
		"--create-filter=escparse,com,parse",
		"--create-filter=purge,com,purge",
		"--create-filter=pinmap,com,pinmap:--rts=cts --dtr=dsr --break=break",
		"--create-filter=linectl,com,lc:--br=remote --lc=remote",
		"--add-filters=0:com",

		"--create-filter=telnet,tcp,telnet:--comport=server --suppress-echo=yes",
		"--create-filter=lsrmap,tcp,lsrmap",
		"--create-filter=pinmap,tcp,pinmap:--cts=cts --dsr=dsr --dcd=dcd --ring=ring",
		"--create-filter=linectl,tcp,lc:--br=local --lc=local",
	)
	if crypt != "" {
		opts = append(opts, crypt)
	}
	hub := exec.Command(hub4com, append(opts,
		"--add-filters=1:tcp",

		// "--use-driver=serial",
		"--octs=off",
		"--ito="+ITO,
		"--ox="+XO,
		"--ix="+XO,
		"--write-limit="+LIMIT,
		`\\.\COM`+serial,

		"--use-driver=tcp",
		port,
	)...)
	hub.Stdout = os.Stdout
	hub.Stderr = os.Stderr
	go func() {
		li.Println(cmd("Run", hub))
		err = srcError(hub.Run())
		PrintOk(cmd("Close", hub), err)
		if err != nil {
			closer.Close()
		}
	}()
	time.Sleep(time.Second)

	if NGROK_AUTHTOKEN == "" {
		planB(Errorf("empty NGROK_AUTHTOKEN"), ":"+port)
		return
	}

	if errNgrok == nil {
		planB(Errorf("found online client: %s", forwardsTo), ":"+port)
		return
	}
	err = nil

	if false {
		ngr = exec.Command(
			"cmd", "/c", "start", // show window of ngrok client for debug
			ngrokBin,
			"tcp",
			port,
		)
		ngr.Env = []string{"NGROK_AUTHTOKEN=" + NGROK_AUTHTOKEN}
		li.Println(cmd("Run", ngr))
		err = srcError(ngr.Run())
		PrintOk(cmd("Close", hub), err)
	} else {
		_ = ngrokBin
		err = run(context.Background(), ":"+port, false)
	}
	if err != nil {
		if strings.Contains(err.Error(), "ERR_NGROK_105") ||
			strings.Contains(err.Error(), "failed to dial ngrok server") {
			planB(err, ":"+port)
			err = nil
		}
	}
}

func watch(dest string) {
	withForwardsTo(dest)
	for {
		time.Sleep(TO)
		if netstat("-a", dest, "") == "" {
			li.Println("no listen ", dest)
			break
		}
	}
}

func planB(err error, dest string) {
	let.Println(err)
	li.Println("LAN mode - режим локальной сети")
	watch(dest)
}

func withForwardsTo(lPort string) (meta string) {
	meta = ifs + lPort
	li.Println(meta)
	return
}
