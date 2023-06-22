package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ngrok/ngrok-api-go/v5"
	"github.com/ngrok/ngrok-api-go/v5/tunnels"
)

const (
	ansiReset = "\u001B[0m"
	ansiRedBG = "\u001B[41m"
	BUG       = ansiRedBG + "Ж" + ansiReset
)

var (
	letf = log.New(os.Stdout, BUG, log.Ltime|log.Lshortfile)
	ltf  = log.New(os.Stdout, " ", log.Ltime|log.Lshortfile)
	let  = log.New(os.Stdout, BUG, log.Ltime)
	lt   = log.New(os.Stdout, " ", log.Ltime)
)

// Get source of code
func src(deep int) (s string) {
	s = string(debug.Stack())
	str := strings.Split(s, "\n")
	if l := len(str); l <= deep {
		deep = l - 1
		for k, v := range str {
			fmt.Println(k, v)
		}
	}
	s = str[deep]
	s = strings.Split(s, " +0x")[0]
	_, s = path.Split(s)
	s += ":"
	return
}

// Wrap source of code and message to error
func Errorf(format string, args ...any) error {
	return fmt.Errorf(src(8)+" %w", fmt.Errorf(format, args...))
}

// Wrap source of code and error to error
func srcError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(src(8)+" %w", err)
}

func Getenv(key, val string) string {
	s := os.Getenv(key)
	if s == "" {
		return val
	}
	return s
}

func ngrokAPI() (publicURL string, forwardsTo string, err error) {
	if NGROK_API_KEY == "" {
		return "", "", Errorf("empty NGROK_API_KEY")
	}

	// construct the api client
	clientConfig := ngrok.NewClientConfig(NGROK_API_KEY)

	// list all online client
	client := tunnels.NewClient(clientConfig)
	iter := client.List(nil)
	err = iter.Err()
	if err != nil {
		return "", "", srcError(err)
	}

	ctx, ca := context.WithTimeout(context.Background(), time.Second*3)
	defer ca()
	for iter.Next(ctx) {
		err = iter.Err()
		if err != nil {
			return "", "", srcError(err)
		}
		if true { //free version allow only one tunnel
			return iter.Item().PublicURL, iter.Item().ForwardsTo, nil
		}
	}
	return "", "", Errorf("not found online client")
}

func PrintOk(s string, err error) {
	if err != nil {
		let.Println(src(8), s, err)
	} else {
		lt.Println(src(8), s, "ok")
	}
}
