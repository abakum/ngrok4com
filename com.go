package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dixonwille/wmenu/v5"
	"github.com/xlab/closer"
	"go.bug.st/serial/enumerator"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	"golang.org/x/sync/errgroup"
)

func com() {
	var (
		forwardsTo string
		ngrokBin   = `..\ngrok\ngrok.exe`
	)

	li.Println("serial server mode - режим сервера порта")
	li.Println(os.Args[0], "[-]serial [port]")
	li.Println(os.Args)

	ngrokBin = filepath.Join(cwd, ngrokBin)

	if len(os.Args) > 1 {
		serial = abs(os.Args[1])
	}

	if len(os.Args) > 2 {
		port = abs(os.Args[2])
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
			if serial == "" {
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

		"--octs=off",
		`\\.\COM`+serial,

		"--use-driver=tcp",
		port,
	)...)
	hub.Stdout = os.Stdout
	hub.Stderr = os.Stderr
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
	time.Sleep(time.Second)

	if NGROK_AUTHTOKEN == "" {
		planB(Errorf("empty NGROK_AUTHTOKEN"))
		return
	}

	_, forwardsTo, err = ngrokAPI(NGROK_API_KEY)
	if err == nil {
		planB(Errorf("found online client: %s", forwardsTo))
		return
	}
	err = nil

	if false {
		ngr := exec.Command(
			"cmd", "/c", "start", // show window of ngrok client for debug
			ngrokBin,
			"tcp",
			port,
		)
		ngr.Env = []string{"NGROK_AUTHTOKEN=" + NGROK_AUTHTOKEN}
		closer.Bind(func() {
			if ngr.Process != nil && ngr.ProcessState == nil {
				PrintOk("ngrok Kill", ngr.Process.Kill())
			}
		})
		err = srcError(ngr.Run())
	} else {
		_ = ngrokBin
		err = run(context.Background(), ":"+port)
	}
	if err != nil {
		if strings.Contains(err.Error(), "ERR_NGROK_105") ||
			strings.Contains(err.Error(), "failed to dial ngrok server") {
			planB(err)
			err = nil
		}
	}
}

func planB(er error) {
	s := "LAN mode - режим локальной сети"
	i := 0
	let.Println(er)
	ifaces, er := net.Interfaces()
	if er == nil {
		for _, ifac := range ifaces {
			addrs, er := ifac.Addrs()
			if er != nil {
				continue
			}
			for _, addr := range addrs {
				if strings.Contains(addr.String(), ":") ||
					strings.HasPrefix(addr.String(), "127.") {
					continue
				}
				s += "\n\t" + addr.String()
				i++
			}
		}
	}
	if i > 0 {
		li.Println(s)
		closer.Hold()
	} else {
		letf.Println("no ifaces for server")
	}
}

// https://github.com/ngrok/ngrok-go/blob/main/examples/ngrok-lite/main.go
func run(ctx context.Context, dest string) error {
	ctxWT, caWT := context.WithTimeout(ctx, time.Second)
	defer caWT()
	sess, er := ngrok.Connect(ctxWT,
		ngrok.WithAuthtoken(NGROK_AUTHTOKEN),
	)
	if er != nil {
		return Errorf("Connect %w", er)
	}
	sess.Close()

	ctx, ca := context.WithCancel(ctx)
	defer func() {
		if er != nil {
			ca()
		}
	}()

	tun, er := ngrok.Listen(ctx,
		config.TCPEndpoint(),
		ngrok.WithAuthtoken(NGROK_AUTHTOKEN),
		ngrok.WithStopHandler(func(ctx context.Context, sess ngrok.Session) error {
			go func() {
				time.Sleep(time.Millisecond * 10)
				ca()
			}()
			return nil
		}),
		ngrok.WithDisconnectHandler(func(ctx context.Context, sess ngrok.Session, err error) {
			PrintOk("WithDisconnectHandler", err)
			if err == nil {
				go func() {
					time.Sleep(time.Millisecond * 10)
					ca()
				}()
			}
		}),
	)
	if er != nil {
		return srcError(er)
	}

	ltf.Println("tunnel created:", tun.URL())

	for {
		conn, er := tun.Accept()
		if er != nil {
			return srcError(er)
		}

		// ltf.Println("accepted connection from", conn.RemoteAddr())

		// go PrintOk("connection closed:", handleConn(ctx, dest, conn))
		go handleConn(ctx, dest, conn)
	}
}

func handleConn(ctx context.Context, dest string, conn net.Conn) error {
	defer conn.Close()
	next, er := net.Dial("tcp", dest)
	if er != nil {
		return srcError(er)
	}
	defer next.Close()

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		_, er := io.Copy(next, conn)
		next.(*net.TCPConn).CloseWrite()
		return srcError(er)
	})
	g.Go(func() error {
		_, er := io.Copy(conn, next)
		return srcError(er)
	})

	return g.Wait()
}
