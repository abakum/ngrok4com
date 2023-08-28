package main

import (
	"embed"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xlab/closer"
	"go.bug.st/serial/enumerator"
)

const (
	EMULATOR = "com0com - serial port emulator"
	BIN      = "bin"
	// BAUD = "19200"
	BAUD = "921600"
)

var (
	//go:embed NGROK_AUTHTOKEN.txt
	NGROK_AUTHTOKEN string
	//go:embed NGROK_API_KEY.txt
	NGROK_API_KEY string
	//go:embed bin/*
	bin embed.FS

	crypt,
	cwd,
	serial string
	port    = "7000"
	hub4com = `hub4com.exe`
	kitty   = `kitty_portable.exe`
	err     error
	opts    = []string{"--baud=" + BAUD}
	ports   []*enumerator.PortDetails
)

func main() {
	defer closer.Close()

	closer.Bind(func() {
		if err != nil {
			let.Println(err)
			defer os.Exit(1)
		}
		// pressEnter()
	})

	NGROK_AUTHTOKEN = Getenv("NGROK_AUTHTOKEN", NGROK_AUTHTOKEN) //if emty then local mode
	// NGROK_AUTHTOKEN += "-"                                       // emulate bad token or no internet
	// NGROK_AUTHTOKEN = ""                                   // emulate local mode
	NGROK_API_KEY = Getenv("NGROK_API_KEY", NGROK_API_KEY) //if emty then no crypt
	// NGROK_API_KEY = ""                                     // emulate no crypt

	if NGROK_API_KEY != "" {
		crypt = "--create-filter=crypt,tcp,crypt:--secret=" + NGROK_API_KEY
	}

	cwd, err = os.Getwd()
	if err != nil {
		err = srcError(err)
		return
	}

	hub4com, err = write(hub4com)
	if err != nil {
		err = srcError(err)
		return
	}

	if len(os.Args) > 1 {
		i, er := strconv.Atoi(abs(os.Args[1]))
		if er != nil || i >= 9600 {
			tty()
			return
		}
	}
	com()
}

func abs(s string) string {
	if strings.HasPrefix(s, "-") {
		NGROK_AUTHTOKEN = "" // no ngrok
		NGROK_API_KEY = ""   // no crypt
		crypt = ""
		return strings.TrimPrefix(s, "-")
	}
	return s
}

func write(fn string) (binFN string, err error) {
	binFN = path.Join(BIN, fn)
	bytes, err := bin.ReadFile(binFN)
	if err != nil {
		err = srcError(err)
		return
	}
	binDir := filepath.Join(cwd, BIN)
	_, err = os.Stat(binDir)
	if err != nil {
		err = os.MkdirAll(binDir, 0666)
		if err != nil {
			err = srcError(err)
			return
		}
	}
	binFN = filepath.Join(binDir, fn)
	_, err = os.Stat(binFN)
	if err == nil {
		return
	}
	ltf.Println(binFN, len(bytes))
	err = os.WriteFile(binFN, bytes, 0666)
	return
}
