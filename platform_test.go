//go:build linux || darwin || windows

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRolloutReaderAllowsRenameAndDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rollout.jsonl")
	if err := os.WriteFile(path, []byte("record\n"), 0600); err != nil {
		t.Fatal(err)
	}
	reader, err := openRollout(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	moved := path + ".moved"
	if err := os.Rename(path, moved); err != nil {
		t.Fatalf("monitor handle prevented rename: %v", err)
	}
	if err := os.Remove(moved); err != nil {
		t.Fatalf("monitor handle prevented deletion: %v", err)
	}
	// The old handle must remain readable until the final drain completes.
	buffer := make([]byte, 7)
	if _, err := reader.ReadAt(buffer, 0); err != nil || string(buffer) != "record\n" {
		t.Fatalf("drain after deletion: data=%q error=%v", buffer, err)
	}
}
