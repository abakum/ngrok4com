package main

import (
	"bytes"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/dixonwille/wmenu/v5"
	"github.com/f1bonacc1/glippy"
	"github.com/xlab/closer"
	"go.bug.st/serial/enumerator"
	"gopkg.in/ini.v1"
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

	if len(os.Args) > 1 {
		_, err = strconv.Atoi(os.Args[1])
		if err != nil {
			// ngrok4com host
			// ngrok4com -host
			host = abs(os.Args[1])
		} else {
			// ngrok4com +
			// ngrok4com baud
			// ngrok4com -
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

	mode := "LAN mode - режим локальной сети"
	if !(crypt == "" || errNgrok != nil || host != "") {
		tcp, err = url.Parse(publicURL)
		if err != nil {
			err = srcError(err)
			return
		}
		host = tcp.Host
		connect, inLAN := fromNgrok(forwardsTo)
		if inLAN == "" || plus {
			mode = "ngrok mode - режим ngrok"
		} else {
			host = connect
		}
	}
	li.Println(mode)

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
		if baud != "" {
			opts = []string{
				"-sercfg",
				baud,
				"-serial",
				"COM" + serial,
			}
			PrintOk("cmdFromClipBoard", command())
			ki = exec.Command(kitty, opts...)

			li.Println(cmd("Run", ki))
			err = srcError(ki.Run())
			PrintOk(cmd("Close", ki), err)
		}
		menu := wmenu.NewMenu("Choose baud and seconds delay for Ctrl-F2 or commands from clipboard case delay>" + DELAY +
			"\nВыбери скорость и задержку в секундах для Ctrl-F2 или команд из буфера обмена если задержка>" + DELAY)
		menu.Action(func(opts []wmenu.Opt) error {
			for _, v := range opts {
				if strings.HasPrefix(v.Text, "0") {
					commandDelay = v.Text
				} else {
					baud = v.Text
				}
			}
			li.Println("baud", baud)
			li.Println("commandDelay", commandDelay)
			return nil
		})
		menu.InitialIndex(0)
		menu.Option(DELAY, 0, commandDelay == DELAY, nil)
		menu.Option("115200", 1, baud == "115200", nil)
		menu.Option("0.2", 2, commandDelay == "0.2", nil)
		menu.Option("38400", 3, baud == "38400", nil)
		menu.Option("0.4", 4, commandDelay == "0.4", nil)
		menu.Option("57600", 5, baud == "57600", nil)
		menu.Option("0.6", 6, commandDelay == "0.6", nil)
		menu.Option("0.7", 7, commandDelay == "0.7", nil)
		menu.Option("0.08", 8, commandDelay == "0.08", nil)
		menu.Option("9600", 9, baud == "9600" || baud == "", nil)
		if !strings.Contains(",115200,38400,57600,9600,,", ","+baud+",") {
			menu.Option(baud, 10, true, nil)
		}
		if menu.Run() != nil {
			return
		}
	}
	// closer.Hold()
}

func SetValue(section *ini.Section, key, val string) (set bool) {
	set = section.Key(key).String() != val
	if set {
		ltf.Println(key, val)
		section.Key(key).SetValue(val)
	}
	return
}

func command() error {
	ini.PrettyFormat = false
	iniFile, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: false,
	}, kittyINI)
	if err != nil {
		return err
	}
	section := iniFile.Section("KiTTY")
	ok := SetValue(section, "commanddelay", commandDelay)
	if ok {
		err = iniFile.SaveTo(kittyINI)
		if err != nil {
			return err
		}
	}
	if commandDelay == DELAY {
		return nil
	}
	text, err := glippy.Get()
	if err != nil {
		return err
	}
	if text == "" {
		return fmt.Errorf("empty ClipBoard")
	}
	temp, err := os.CreateTemp("", "cmdFromClipBoard")
	if err != nil {
		return err
	}
	clip := temp.Name()
	defer os.Remove(clip)
	n, err := temp.WriteString(text)
	if err != nil {
		return err
	}
	if n != len(text) {
		return fmt.Errorf("error write ClipBoard to %s", clip)
	}
	opts = append(opts,
		"-cmd",
		clip,
	)
	return nil
}

func contains(net, ip string) bool {
	network, err := netip.ParsePrefix(net)
	if err != nil {
		return false
	}
	ipContains, err := netip.ParsePrefix(ip)
	if err != nil {
		return false
	}
	return network.Contains(ipContains.Addr())
}

func fromNgrok(forwardsTo string) (connect, inLAN string) {
	netsPorts := strings.Split(forwardsTo, ":")
	nets := strings.Split(netsPorts[0], ",")
	for _, ip := range ips {
		for _, net := range nets {
			listen := strings.Split(net, "/")[0]
			if !contains(net, ip) {
				continue
			}
			inLAN = listen
		}
	}
	if len(netsPorts) > 1 {
		port = netsPorts[1]
	}
	connect = fmt.Sprintf("%s:%s", inLAN, port)
	return
}
