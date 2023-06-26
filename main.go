package main

import (
	_ "embed"
	"os"
	"strconv"
	"strings"
)

const (
	emulator = "com0com - serial port emulator"
)

var (
	//go:embed NGROK_AUTHTOKEN.txt
	NGROK_AUTHTOKEN string
	//go:embed NGROK_API_KEY.txt
	NGROK_API_KEY string

	crypt   = "--data=8" //placeholder
	port    = "7000"
	hub4com = `..\hub4com\hub4com.exe`
)

func main() {
	NGROK_AUTHTOKEN = Getenv("NGROK_AUTHTOKEN", NGROK_AUTHTOKEN) //if emty then local mode
	// NGROK_AUTHTOKEN += "-"                                       // emulate bad token or no internet
	// NGROK_AUTHTOKEN = ""                                   // emulate local mode
	NGROK_API_KEY = Getenv("NGROK_API_KEY", NGROK_API_KEY) //if emty then no crypt
	// NGROK_API_KEY = ""                                     // emulate no crypt
	if len(os.Args) > 1 {
		i, err := strconv.Atoi(os.Args[1])
		if err != nil || i >= 9600 {
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
		return strings.TrimPrefix(s, "-")
	}
	return s
}
