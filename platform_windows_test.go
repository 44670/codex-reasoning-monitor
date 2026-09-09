package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func acquireTestLock(file *os.File) error {
	return windows.LockFileEx(windows.Handle(file.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &windows.Overlapped{})
}

func stopTestProcess(process *os.Process) error {
	// Go cannot send os.Interrupt to a Windows child. Console Ctrl+C is a manual
	// platform check; use TerminateProcess to reliably release fixture resources.
	return process.Kill()
}
