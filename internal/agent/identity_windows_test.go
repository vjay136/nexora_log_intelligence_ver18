//go:build windows

package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsIdentityIsStableAcrossAppend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.log")
	if err := os.WriteFile(path, []byte("first\n"), 0600); err != nil {
		t.Fatal(err)
	}
	readIdentity := func() (uint64, uint64) {
		t.Helper()
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			t.Fatal(err)
		}
		device, inode, err := identity(file, info)
		if err != nil {
			t.Fatal(err)
		}
		return device, inode
	}

	deviceBefore, inodeBefore := readIdentity()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.WriteString("second\n"); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	deviceAfter, inodeAfter := readIdentity()
	if deviceAfter != deviceBefore || inodeAfter != inodeBefore {
		t.Fatalf("identity changed after append: (%d,%d) -> (%d,%d)", deviceBefore, inodeBefore, deviceAfter, inodeAfter)
	}
}
