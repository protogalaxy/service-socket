// Copyright (C) 2015 The Protogalaxy Project
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
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
	w := socket.NewMessageWriter(nil, nil)
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
	w := socket.NewMessageWriter(&writer, make(chan []byte))
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

type WriterError struct{}

func (w *WriterError) Write(b []byte) (int, error) {
	return 0, errors.New("err")
}

func TestMessageWriterExitOnWriteError(t *testing.T) {
	t.Parallel()
	writer := &WriterError{}
	w := socket.NewMessageWriter(writer, make(chan []byte))
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
