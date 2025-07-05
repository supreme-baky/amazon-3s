package info

import (
	"fmt"
	"os"
)

func PrintInfo() {
	fmt.Println(`Simple Storage Service

Usage:
  triple-s [-port <N>] [-dir <S>]
  triple-s --help

Options:
  --help       Show this help screen
  --port N     Port number to listen on (default 8080)
  --dir S      Base URL (default "http://localhost")`)
	os.Exit(0)
}
