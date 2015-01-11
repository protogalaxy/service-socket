package socket_test

import (
	"testing"
	"time"

	"github.com/protogalaxy/service-socket/socket"
)

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

func socketSendMessage(t *testing.T, msgs chan<- socket.Message, socketId socket.Id, data string) {
	select {
	case msgs <- socket.Message{
		SocketId: socketId,
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
