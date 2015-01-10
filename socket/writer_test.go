package socket_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/protogalaxy/service-socket/socket"
)

type CloserMock struct {
	CloseFunc func() error
}

func (m *CloserMock) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func TestMessageWriterClosesReader(t *testing.T) {
	t.Parallel()
	w := socket.NewMessageWriter(nil)
	var called bool
	w.Reader = &CloserMock{
		CloseFunc: func() error {
			called = true
			return nil
		},
	}
	done := make(chan struct{})
	go func() {
		w.Run()
		close(done)
	}()
	w.Close()
	select {
	case <-done:
		if !called {
			t.Fatal("Reader should be closed")
		}
	case <-time.After(time.Millisecond):
		t.Fatal("Writer not closing")
	}
}

func TestMessageWriterMessagesAreWritten(t *testing.T) {
	t.Parallel()
	var writer bytes.Buffer
	w := socket.NewMessageWriter(&writer)
	done := make(chan struct{})
	go func() {
		w.Run()
		close(done)
	}()
	sendMessage(t, w.Messages(), "abc")
	sendMessage(t, w.Messages(), "d")
	w.Close()
	select {
	case <-done:
		if writer.String() != "abcd" {
			t.Fatalf("Expecting 'abcd' to be written but got '%s'", writer.String())
		}
	case <-time.After(time.Millisecond):
		t.Fatal("Writer not closing")
	}
}

func TestMessagesAreDiscarded(t *testing.T) {
	t.Parallel()
	var writer bytes.Buffer
	w := socket.NewMessageWriter(&writer)
	done := make(chan struct{})
	go func() {
		w.Run()
		close(done)
	}()
	dc := make(chan struct{})
	discardDone := make(chan struct{})
	go func() {
		w.DiscardUntil(dc)
		close(discardDone)
	}()
	select {
	case <-done:
		sendMessage(t, w.Messages(), "msg")
		close(dc)
		select {
		case <-discardDone:
			if writer.String() != "" {
				t.Fatalf("No messages should be written but got '%s'", writer.String())
			}
		case <-time.After(time.Millisecond):
			t.Fatal("Discard not finishing")
		}
	case <-time.After(time.Millisecond):
		t.Fatal("Writer was not closed")
	}
}

type WriterError struct{}

func (w *WriterError) Write(b []byte) (int, error) {
	return 0, errors.New("err")
}

func TestMessageWriterExitOnWriteError(t *testing.T) {
	t.Parallel()
	writer := &WriterError{}
	w := socket.NewMessageWriter(writer)
	done := make(chan struct{})
	go func() {
		w.Run()
		close(done)
	}()
	sendMessage(t, w.Messages(), "abc")
	select {
	case <-done:
	case <-time.After(time.Millisecond):
		t.Fatal("Writer should exit after write error")
	}
}

func sendMessage(t *testing.T, m chan<- []byte, msg string) {
	select {
	case m <- []byte(msg):
	case <-time.After(time.Millisecond):
		t.Fatal("Message not consumed")
	}
}
