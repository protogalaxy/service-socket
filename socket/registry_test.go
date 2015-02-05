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
	"testing"
	"time"

	"github.com/protogalaxy/service-socket/socket"
)

func TestSocketIDTOString(t *testing.T) {
	sid := socket.ID(123456)
	if s := sid.String(); s != "1e240" {
		t.Errorf("Socket ID should be converted to string but got: %s", s)
	}
}

func TestSocketParseID(t *testing.T) {
	sid := socket.ID(123456)
	if psid, err := socket.ParseID(sid.String()); err != nil || sid != psid {
		t.Errorf("Socket id should be parsed from string but got: %d", psid)
	}
}

func TestSocketRegistrySendMessage(t *testing.T) {
	t.Parallel()
	reg := socket.NewRegistry()
	done := make(chan struct{})
	go func() {
		reg.Run()
		close(done)
	}()

	c1 := make(chan []byte, 1)
	id1, err := reg.Register(c1)
	if err != nil {
		t.Errorf("Registering socket should not fail but got: %s", err)
	}

	c2 := make(chan []byte, 1)
	id2, err := reg.Register(c2)
	if err != nil {
		t.Errorf("Registering socket should not fail but got: %s", err)
	}

	socketSendMessage(t, reg.Messages(), id1, "abc")
	socketSendMessage(t, reg.Messages(), id2, "def")

	checkReceivedMessage(t, c1, "abc")
	checkReceivedMessage(t, c2, "def")

	reg.Close()
	select {
	case <-done:
	case <-time.After(time.Millisecond):
		t.Fatal("Registry not closing")
	}
}

func TestSocketRegistryMessageQueueFull(t *testing.T) {
	t.Parallel()
	reg := socket.NewRegistry()
	done := make(chan struct{})
	go func() {
		reg.Run()
		close(done)
	}()

	c1 := make(chan []byte)
	id1, err := reg.Register(c1)
	if err != nil {
		t.Errorf("Registering socket should not fail but got: %s", err)
	}

	socketSendMessage(t, reg.Messages(), id1, "abc")

	time.Sleep(time.Millisecond) // Wait for the send to be processed

	select {
	case m := <-c1:
		t.Errorf("No messages should be received but got: %s", m)
	case <-time.After(time.Millisecond):
	}

	reg.Close()
	select {
	case <-done:
	case <-time.After(time.Millisecond):
		t.Fatal("Registry not closing")
	}
}

func socketSendMessage(t *testing.T, msgs chan<- socket.Message, socketId socket.ID, data string) {
	select {
	case msgs <- socket.Message{
		SocketID: socketId,
		Data:     []byte(data),
	}:
	case <-time.After(time.Millisecond):
		t.Fatalf("Message to registered channel (%d) not sent", socketId)
	}
}

func checkReceivedMessage(t *testing.T, msgs <-chan []byte, data string) {
	select {
	case m := <-msgs:
		if string(m) != data {
			t.Errorf("Expecting to receive '%s' but got '%s'", data, string(m))
		}
	case <-time.After(time.Millisecond):
		t.Fatal("No message in channel")
	}
}

func TestSocketRegistryUnregister(t *testing.T) {
	t.Parallel()
	reg := socket.NewRegistry()
	done := make(chan struct{})
	go func() {
		reg.Run()
		close(done)
	}()

	c1 := make(chan []byte)
	id1, err := reg.Register(c1)
	if err != nil {
		t.Errorf("Registering socket should not fail but got: %s", err)
	}

	reg.Unregister(id1)

	// Sending on a nonexistent socket should drop the message.
	socketSendMessage(t, reg.Messages(), id1, "")

	reg.Close()
	select {
	case <-done:
	case <-time.After(time.Millisecond):
		t.Fatal("Registry not closing")
	}
}
