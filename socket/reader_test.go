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
	"io"
	"testing"
	"time"

	"github.com/protogalaxy/service-socket/socket"
)

type MockReader struct {
	r io.Reader
}

func (r *MockReader) ReadMessage() ([]byte, error) {
	data := make([]byte, 2)
	_, err := r.r.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func TestMessageReaderMessagesAreRead(t *testing.T) {
	t.Parallel()
	reader := MockReader{bytes.NewReader([]byte("abcdef"))}
	r := socket.NewMessageReader(&reader)
	done := make(chan struct{})
	go func() {
		r.Run()
		close(done)
	}()
	if m := readMessage(t, r.Messages()); m != "ab" {
		t.Fatalf("Expecting message 'ab' but got '%s'", m)
	}
	if m := readMessage(t, r.Messages()); m != "cd" {
		t.Fatalf("Expecting message 'cd' but got '%s'", m)
	}
	r.Close()
	select {
	case <-done:
	case <-time.After(time.Millisecond):
		t.Fatal("Reader not closing")
	}
}

func TestMessageReaderExitOnReadError(t *testing.T) {
	t.Parallel()
	reader := MockReader{bytes.NewReader([]byte(""))}
	r := socket.NewMessageReader(&reader)
	done := make(chan struct{})
	go func() {
		r.Run()
		close(done)
	}()
	select {
	case data := <-r.Messages():
		t.Fatalf("No messages should be sent but got '%s'", string(data))
	case <-done:
	case <-time.After(time.Millisecond):
		t.Fatal("Reader shouled be closed on error")
	}
}

func TestMessageReaderWriterIsClosed(t *testing.T) {
	t.Parallel()
	reader := MockReader{bytes.NewReader([]byte(""))}
	r := socket.NewMessageReader(&reader)
	var called bool
	r.Writer = &CloserMock{
		CloseFunc: func() error {
			called = true
			return nil
		},
	}
	done := make(chan struct{})
	go func() {
		r.Run()
		close(done)
	}()
	select {
	case <-done:
		if !called {
			t.Fatal("Writer should be closed")
		}
	case <-time.After(time.Millisecond):
		t.Fatal("Reader shouled be closed on error")
	}
}

func readMessage(t *testing.T, m <-chan []byte) string {
	select {
	case data := <-m:
		return string(data)
	case <-time.After(time.Millisecond):
		t.Fatal("No messages to read")
		return ""
	}
}
