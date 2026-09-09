package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func openRollout(path string) (*os.File, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	// Go's os.Open does not request FILE_SHARE_DELETE. A long-lived reader must
	// allow Codex to rename/delete the rollout while this handle is still open.
	handle, err := windows.CreateFile(name, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(handle), path), nil
}

func prepareConsole(out *os.File) (restore func(), supported bool) {
	handle := windows.Handle(out.Fd())
	var mode uint32
	if windows.GetConsoleMode(handle, &mode) != nil {
		return nil, false // Redirected output: --color always can still emit ANSI.
	}
	if windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING) != nil {
		return nil, false
	}
	return func() { _ = windows.SetConsoleMode(handle, mode) }, true
}
