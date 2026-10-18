//go:build linux

package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInputWatcherWakesForAppendAndCoalesces(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "events.log")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	wake, closeWatcher := watchInputDirectories(ctx, []InputConfig{{Path: path}})
	defer closeWatcher()
	defer cancel()

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("one\ntwo\n"); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-wake:
	case <-time.After(2 * time.Second):
		t.Fatal("append did not wake input scan")
	}

	// The wake channel has capacity one. Burst writes must not queue one task
	// per filesystem event or block the watcher goroutine.
	for index := 0; index < 32; index++ {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		_, writeErr := file.WriteString("burst\n")
		closeErr := file.Close()
		if writeErr != nil {
			t.Fatal(writeErr)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	select {
	case <-wake:
	case <-time.After(2 * time.Second):
		t.Fatal("burst appends did not wake input scan")
	}
	select {
	case <-wake:
	default:
	}
}

func TestInputWatcherSetupFailureLeavesReconciliationAvailable(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing", "events.log")
	wake, closeWatcher := watchInputDirectories(context.Background(), []InputConfig{{Path: missingPath}})
	defer closeWatcher()
	if wake == nil {
		t.Fatal("watch setup failure must return a usable wake channel")
	}
	select {
	case <-wake:
		t.Fatal("failed watcher unexpectedly emitted a wake")
	default:
	}
}
