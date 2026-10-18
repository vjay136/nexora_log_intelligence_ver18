package output

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type queuedTestWriter struct {
	entered chan struct{}
	release chan struct{}
	mu      sync.Mutex
	writes  []string
	err     error
}

func (w *queuedTestWriter) Write(data []byte, route string) error {
	if w.entered != nil {
		w.entered <- struct{}{}
		<-w.release
	}
	w.mu.Lock()
	w.writes = append(w.writes, string(data))
	w.mu.Unlock()
	return w.err
}
func (w *queuedTestWriter) Flush() error { return nil }
func (w *queuedTestWriter) Close() error { return nil }

func TestRouterDestinationIsolationAndFlushBarrier(t *testing.T) {
	slow := &queuedTestWriter{entered: make(chan struct{}, 2), release: make(chan struct{}, 2)}
	fast := &queuedTestWriter{}
	r, err := NewRouter([]Config{
		{Type: "captured", Writer: slow, QueueCapacity: 1},
		{Type: "captured", Writer: fast, QueueCapacity: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if err := r.Write([]byte("first"), "normal"); err != nil {
		t.Fatal(err)
	}
	<-slow.entered
	if err := r.Write([]byte("second"), "normal"); err != nil {
		t.Fatal(err)
	}
	fast.mu.Lock()
	got := len(fast.writes)
	fast.mu.Unlock()
	deadline := time.After(time.Second)
	for got != 2 {
		select {
		case <-deadline:
			t.Fatal("fast destination stalled behind slow destination")
		default:
			time.Sleep(time.Millisecond)
		}
		fast.mu.Lock()
		got = len(fast.writes)
		fast.mu.Unlock()
	}
	blocked := make(chan error, 1)
	go func() { blocked <- r.Write([]byte("third"), "normal") }()
	select {
	case <-blocked:
		t.Fatal("queue did not bound pending writes")
	case <-time.After(20 * time.Millisecond):
	}
	slow.release <- struct{}{}
	<-slow.entered
	select {
	case err := <-blocked:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("bounded writer did not resume")
	}
	flushed := make(chan error, 1)
	go func() { flushed <- r.Flush() }()
	select {
	case <-flushed:
		t.Fatal("flush passed pending slow writes")
	case <-time.After(20 * time.Millisecond):
	}
	slow.release <- struct{}{}
	<-slow.entered
	slow.release <- struct{}{}
	select {
	case err := <-flushed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("flush did not finish")
	}
}

func TestRouterFlushReportsWorkerFailure(t *testing.T) {
	want := errors.New("write failed")
	w := &queuedTestWriter{err: want}
	r, err := NewRouter([]Config{{Type: "captured", Writer: w}})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if err := r.Write([]byte("x"), "normal"); err != nil {
		t.Fatal(err)
	}
	if err := r.Flush(); !errors.Is(err, want) {
		t.Fatalf("Flush error = %v, want %v", err, want)
	}
}

func TestRouterRejectsUnboundedQueueCapacity(t *testing.T) {
	if _, err := NewRouter([]Config{{Type: "discard", QueueCapacity: 65537}}); err == nil {
		t.Fatal("unbounded queue capacity accepted")
	}
}
