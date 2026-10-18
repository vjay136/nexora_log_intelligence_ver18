//go:build linux

package agent

import (
	"context"
	"path/filepath"
	"sync"

	"golang.org/x/sys/unix"
)

// watchInputDirectories coalesces filesystem notifications into bounded wake
// signals. The agent still runs periodic reconciliation scans independently.
func watchInputDirectories(ctx context.Context, inputs []InputConfig) (<-chan struct{}, func()) {
	wake := make(chan struct{}, 1)
	watcher, err := newInputWatcher(inputs)
	if err != nil || watcher == nil {
		return wake, func() {}
	}
	watcher.start(ctx, wake)
	return wake, watcher.close
}

type inputWatcher struct {
	fd      int
	closeMu sync.Mutex
	closed  bool
}

func newInputWatcher(inputs []InputConfig) (*inputWatcher, error) {
	fd, err := unix.InotifyInit1(unix.IN_CLOEXEC | unix.IN_NONBLOCK)
	if err != nil {
		return nil, err
	}
	watcher := &inputWatcher{fd: fd}
	seen := make(map[string]struct{}, len(inputs))
	for _, input := range inputs {
		directory := filepath.Dir(input.Path)
		if _, ok := seen[directory]; ok {
			continue
		}
		seen[directory] = struct{}{}
		if _, err := unix.InotifyAddWatch(fd, directory, unix.IN_MODIFY|unix.IN_CLOSE_WRITE|unix.IN_MOVED_TO|unix.IN_CREATE|unix.IN_ATTRIB|unix.IN_DELETE|unix.IN_MOVE_SELF|unix.IN_DELETE_SELF); err != nil {
			watcher.close()
			return nil, err
		}
	}
	return watcher, nil
}

func (watcher *inputWatcher) start(ctx context.Context, wake chan<- struct{}) {
	go func() {
		buffer := make([]byte, 64*1024)
		for {
			watcher.closeMu.Lock()
			closed := watcher.closed
			watcher.closeMu.Unlock()
			if closed || ctx.Err() != nil {
				return
			}
			n, err := unix.Read(watcher.fd, buffer)
			if err == nil && n > 0 {
				// Treat every event batch, including IN_Q_OVERFLOW, as a cue to
				// reconcile actual file contents instead of trusting notifications.
				select {
				case wake <- struct{}{}:
				default:
				}
				continue
			}
			if err != nil && err != unix.EAGAIN && err != unix.EINTR {
				return // periodic reconciliation remains active
			}
			unix.Poll([]unix.PollFd{{Fd: int32(watcher.fd), Events: unix.POLLIN}}, 250)
		}
	}()
}

func (watcher *inputWatcher) close() {
	watcher.closeMu.Lock()
	defer watcher.closeMu.Unlock()
	if watcher.closed {
		return
	}
	watcher.closed = true
	_ = unix.Close(watcher.fd)
}
