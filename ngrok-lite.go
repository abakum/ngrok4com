package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
	"time"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	"golang.org/x/sync/errgroup"
)

func cmd(s string, c *exec.Cmd) string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf(`%s "%s" %s`, s, c.Args[0], strings.Join(c.Args[1:], " "))
}

func ns(a string) string {
	var (
		err     error
		bBuffer bytes.Buffer
	)
	opts := []string{
		"-n",
		"-p",
		"TCP",
		"-o",
	}
	if a != "" {
		opts = append(opts, a)
	}
	stat := exec.Command("netstat", opts...)
	stat.Stdout = &bBuffer
	stat.Stderr = &bBuffer
	err = stat.Run()
	if err != nil {
		PrintOk(cmd("Run", stat), err)
		return ""
	}
	return bBuffer.String()
}

func nStat(all, a, host, pid string) (contains string) {
	var (
		err     error
		bBuffer bytes.Buffer
	)
	ok := "LISTENING"
	if a == "" {
		ok = "ESTABLISHED"
	}
	bBuffer.WriteString(all)
	for {
		contains, err = bBuffer.ReadString('\n')
		if err != nil {
			return ""
		}
		if strings.Contains(contains, host) && strings.Contains(contains, ok) && strings.Contains(contains, pid) {
			return
		}
	}
}

func netstat(a, host, pid string) (contains string) {
	return nStat(ns(a), a, host, pid)
}

// https://github.com/ngrok/ngrok-go/blob/main/examples/ngrok-lite/main.go
func run(ctx context.Context, dest string, http bool) error {
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
	endpoint := config.TCPEndpoint(config.WithForwardsTo(withForwardsTo(dest)))
	if http {
		endpoint = config.HTTPEndpoint(config.WithForwardsTo(withForwardsTo(dest)))
	}
	tun, err := ngrok.Listen(ctx,
		endpoint,
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
	if err != nil {
		return srcError(err)
	}

	ltf.Println("tunnel created:", tun.URL())

	for {
		conn, err := tun.Accept()
		if err != nil {
			return srcError(err)
		}
		ltf.Println("accepted connection from", conn.RemoteAddr(), "to", conn.LocalAddr())

		next, err := net.Dial("tcp", dest)
		if err != nil {
			return srcError(err)
		}

		PrintOk("connection closed", handleConn(ctx, next, conn))
	}
}

func handleConn(ctx context.Context, next, conn net.Conn) error {
	defer conn.Close()
	defer next.Close()

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		_, err := io.Copy(next, conn)
		next.(*net.TCPConn).CloseWrite() //for close without error
		time.Sleep(time.Millisecond * 7)
		next.Close()
		return srcError(err)
	})
	g.Go(func() error {
		_, err := io.Copy(conn, next)
		conn.Close()
		return srcError(err)
	})

	return g.Wait()
}
