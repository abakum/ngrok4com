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
	"time"

	"github.com/xlab/closer"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	"golang.org/x/sync/errgroup"
)

var (
	//go:embed NGROK_AUTHTOKEN.txt
	NGROK_AUTHTOKEN string
)

func main() {
	var (
		err     error
		hub4com = `..\hub4com\hub4com.exe`
		com     = "7"
		port    = "7000"
	)
	defer closer.Close()

	closer.Bind(func() {
		if err != nil {
			let.Println(err)
			defer os.Exit(1)
		}
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
	}

	_, forwardsTo, err := ngrokAPI()
	if err == nil {
		err = Errorf("found online client: %s", forwardsTo)
		return
	}
	err = nil

	hub := exec.Command(
		hub4com,
		"--create-filter=escparse,com,parse",
		"--create-filter=purge,com,purge",
		"--create-filter=pinmap,com,pinmap:--rts=cts --dtr=dsr --break=break",
		"--create-filter=linectl,com,lc:--br=remote --lc=remote",
		"--add-filters=0:com",
		"--create-filter=telnet,tcp,telnet:--comport=server --suppress-echo=yes",
		"--create-filter=lsrmap,tcp,lsrmap",
		"--create-filter=pinmap,tcp,pinmap:--cts=cts --dsr=dsr --dcd=dcd --ring=ring",
		"--create-filter=linectl,tcp,lc:--br=local --lc=local",
		"--add-filters=1:tcp",
		"--octs=off",
		"COM"+com,
		"--use-driver=tcp",
		"*"+port,
	)
	hub.Stdout = os.Stdout
	hub.Stderr = os.Stderr
	closer.Bind(func() {
		PrintOk("hub4com", hub.Process.Kill())
	})
	go func() {
		err = hub.Run()
		if err != nil {
			letf.Println(err)
			closer.Close()
		}
	}()
	time.Sleep(time.Second)
	PrintOk("ngrok", run(context.Background(), "127.0.0.1:"+port))
}

// https://github.com/ngrok/ngrok-go/blob/main/examples/ngrok-lite/main.go
func run(ctx context.Context, dest string) error {
	tun, err := ngrok.Listen(ctx,
		config.TCPEndpoint(),
		ngrok.WithAuthtoken(Getenv("NGROK_AUTHTOKEN", NGROK_AUTHTOKEN)),
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

		ltf.Println("accepted connection from", conn.RemoteAddr())

		go PrintOk("connection closed:", handleConn(ctx, dest, conn))
	}
}

func handleConn(ctx context.Context, dest string, conn net.Conn) error {
	next, err := net.Dial("tcp", dest)
	if err != nil {
		return srcError(err)
	}

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		_, err := io.Copy(next, conn)
		return srcError(err)
	})
	g.Go(func() error {
		_, err := io.Copy(conn, next)
		return srcError(err)
	})

	return g.Wait()
}
