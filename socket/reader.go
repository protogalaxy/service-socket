package socket

import (
	"io"
	"sync"

	"github.com/golang/glog"
)

type Reader interface {
	ReadMessage() ([]byte, error)
}

// MessageReader is a worker that reads from a specified Reader. Every message read
// is send over its messages channel.
// If the Writer is set it will be closed after the Run terminates.
// The Writer should not be set while the reader is running.
type MessageReader struct {
	r         Reader
	messages  chan []byte
	close     chan struct{}
	closeOnce sync.Once
	Writer    io.Closer
}

// NewMessageReader constructs a new MessageReader for a given Reader.
func NewMessageReader(r Reader) *MessageReader {
	return &MessageReader{
		r:        r,
		messages: make(chan []byte),
		close:    make(chan struct{}),
	}
}

// Run reads messages from the underlying Reader and sends them to the messages channel.
// Method terminates if a read error occurs or the reader is explicitly closed.
// If set the Writer is closed before returning.
func (r *MessageReader) Run() {
	defer func() {
		if r.Writer != nil {
			r.Writer.Close()
		}
	}()
	for {
		glog.V(4).Info("Reading message")
		data, err := r.r.ReadMessage()
		if err != nil {
			glog.Warning("Unable to read: ", err)
			return
		}
		select {
		case r.messages <- data:
			continue
		case <-r.close:
			glog.Info("Closing reader")
			return
		}
	}
}

// Messages returns a receive only channel of read messages.
func (r *MessageReader) Messages() <-chan []byte {
	return r.messages
}

// Close implements a Closer interface.
// The method is idempotents and can be called multiple times.
func (r *MessageReader) Close() error {
	r.closeOnce.Do(func() { close(r.close) })
	return nil
}
