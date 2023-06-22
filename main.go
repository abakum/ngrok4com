package main

import (
	"context"
	_ "embed"
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

var (
	//go:embed NGROK_AUTHTOKEN.txt
	NGROK_AUTHTOKEN string
	//go:embed NGROK_API_KEY.txt
	NGROK_API_KEY string
)

func main() {
	var (
		err      error
		hub4com  = `..\hub4com\hub4com.exe`
		ngrokBin = `..\ngrok\ngrok.exe`
		com      = "7"
		port     = "7000"
		crypt    = "--data=8"
	)
	NGROK_AUTHTOKEN = Getenv("NGROK_AUTHTOKEN", NGROK_AUTHTOKEN) //if emty then local mode
	// NGROK_AUTHTOKEN += "-"                                       // emulate bad token or no internet
	// NGROK_AUTHTOKEN = ""                                   // emulate local mode
	NGROK_API_KEY = Getenv("NGROK_API_KEY", NGROK_API_KEY) //if emty then no crypt
	// NGROK_API_KEY = ""                                     // emulate no crypt
	if NGROK_API_KEY != "" {
		crypt = "--create-filter=crypt,tcp,crypt:--secret=" + NGROK_API_KEY
	}
	defer closer.Close()

	closer.Bind(func() {
		if err != nil {
			let.Println(err)
			defer os.Exit(1)
		}
		time.Sleep(time.Second)
		fmt.Print("Press Enter>")
		fmt.Scanln()
	})

	if len(os.Args) > 1 {
		com = os.Args[1]
	}

	if len(os.Args) > 2 {
		port = os.Args[2]
	}

	cwd, err := os.Getwd()
	if err == nil {
		hub4com = filepath.Join(cwd, hub4com)
		ngrokBin = filepath.Join(cwd, ngrokBin)
	}

	menu := wmenu.NewMenu("Choose serial port")
	menu.Action(func(opts []wmenu.Opt) error {
		com = opts[0].Value.(string)
		return nil
	})
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		err = srcError(err)
		return
	}
	for _, sPort := range ports {
		s := fmt.Sprintf("%s %s", sPort.Name, sPort.Product)
		lt.Println(s)
		if !strings.Contains(sPort.Product, "com0com - serial port emulator") {
			suff := strings.TrimPrefix(sPort.Name, "COM")
			menu.Option(s, suff, suff == com, nil)
		}
	}
	if len(os.Args) < 2 {
		err = menu.Run()
		if err != nil {
			err = srcError(err)
			return
		}
	}

	hub := exec.Command(
		hub4com,
		// "--interface=127.0.0.1",
		"--baud=460800",

		"--create-filter=escparse,com,parse",
		"--create-filter=purge,com,purge",
		"--create-filter=pinmap,com,pinmap:--rts=cts --dtr=dsr --break=break",
		"--create-filter=linectl,com,lc:--br=remote --lc=remote",
		"--add-filters=0:com",

		"--create-filter=telnet,tcp,telnet:--comport=server --suppress-echo=yes",
		"--create-filter=lsrmap,tcp,lsrmap",
		"--create-filter=pinmap,tcp,pinmap:--cts=cts --dsr=dsr --dcd=dcd --ring=ring",
		"--create-filter=linectl,tcp,lc:--br=local --lc=local",
		crypt,
		"--add-filters=1:tcp",

		"--octs=off",
		`\\.\COM`+com,

		"--use-driver=tcp",
		port,
	)
	hub.Stdout = os.Stdout
	hub.Stderr = os.Stderr
	closer.Bind(func() {
		if hub.Process != nil {
			PrintOk("hub4com", hub.Process.Kill())
		}
	})
	go func() {
		err = hub.Run()
		if err != nil {
			letf.Println(err)
			closer.Close()
		}
	}()
	time.Sleep(time.Second)

	if NGROK_AUTHTOKEN == "" {
		planB(Errorf("empty NGROK_AUTHTOKEN"))
		return
	}

	_, forwardsTo, err := ngrokAPI()
	if err == nil {
		planB(Errorf("found online client: %s", forwardsTo))
		return
	}
	err = nil

	if false {
		ngr := exec.Command(
			// "cmd", "/c", "start", // show window of ngrok client for debug
			ngrokBin,
			"tcp",
			port,
		)
		ngr.Env = []string{"NGROK_AUTHTOKEN=" + NGROK_AUTHTOKEN}
		closer.Bind(func() {
			if ngr.Process != nil {
				PrintOk("ngrok", ngr.Process.Kill())
			}
		})
		err = srcError(ngr.Run())
	} else {
		_ = ngrokBin
		err = run(context.Background(), ":"+port)
	}
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") ||
			strings.Contains(err.Error(), "ERR_NGROK_105") {
			planB(err)
			err = nil
		}
	}
}

func planB(err error) {
	defer closer.Hold()
	let.Println(err)
	lt.Println("Plan B: say IP for connect tty without internet")
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return
		}
		for _, addr := range addrs {
			if strings.HasPrefix(addr.String(), "::") {
				continue
			}
			if strings.HasPrefix(addr.String(), "127.") {
				continue
			}
			lt.Println(addr)
		}
	}
}

// https://github.com/ngrok/ngrok-go/blob/main/examples/ngrok-lite/main.go
func run(ctx context.Context, dest string) error {
	ctxWT, caWT := context.WithTimeout(ctx, time.Second)
	defer caWT()
	sess, err := ngrok.Connect(ctxWT,
		ngrok.WithAuthtoken(NGROK_AUTHTOKEN),
	)
	if err != nil {
		return Errorf("Connect %w", err)
	}
	sess.Close()

	ctx, ca := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			ca()
		}
	}()

	tun, err := ngrok.Listen(ctx,
		config.TCPEndpoint(),
		ngrok.WithAuthtoken(NGROK_AUTHTOKEN),
		ngrok.WithStopHandler(func(ctx context.Context, sess ngrok.Session) error {
			go func() {
				time.Sleep(time.Millisecond * 10)
				ca()
			}()
			return nil
		}),
	)
	if err != nil {
		return srcError(err)
	}

	ltf.Println("tunnel created:", tun.URL())

	for {
		conn, err := tun.Accept()
		if err != nil {
			return srcError(err)
		}

		// ltf.Println("accepted connection from", conn.RemoteAddr())

		// go PrintOk("connection closed:", handleConn(ctx, dest, conn))
		go handleConn(ctx, dest, conn)
	}
}

func handleConn(ctx context.Context, dest string, conn net.Conn) error {
	defer conn.Close()
	next, err := net.Dial("tcp", dest)
	if err != nil {
		return srcError(err)
	}
	defer next.Close()

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		_, err := io.Copy(next, conn)
		next.(*net.TCPConn).CloseWrite()
		return srcError(err)
	})
	g.Go(func() error {
		_, err := io.Copy(conn, next)
		return srcError(err)
	})

	return g.Wait()
}

// func handleConn_(ctx context.Context, dest string, conn net.Conn) error {
// 	defer conn.Close()
// 	dial, err := net.Dial("tcp", dest)
// 	if err != nil {
// 		return srcError(err)
// 	}
// 	defer dial.Close()

// 	done := make(chan error, 2)

// 	go func() {
// 		_, err = io.Copy(dial, conn)
// 		// Signal peer that no more data is coming.
// 		dial.(*net.TCPConn).CloseWrite()
// 		// PrintOk("dial<conn", err)
// 		done <- srcError(err)
// 	}()
// 	go func() {
// 		_, err = io.Copy(conn, dial)
// 		// PrintOk("conn<dial", err)
// 		done <- srcError(err)
// 	}()
// 	return <-done
// }
