//go:build linux || darwin

package main

import (
	"os"

	"golang.org/x/term"
)

func openRollout(path string) (*os.File, error) {
	return os.Open(path)
}

func prepareConsole(out *os.File) (restore func(), supported bool) {
	return nil, term.IsTerminal(int(out.Fd()))
}
