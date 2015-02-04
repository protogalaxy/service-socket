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
package socket

import (
	"io"
	"sync"

	"github.com/golang/glog"
)

// Reader is the interface that provides reading of individual messages.
// Every message is read as a whole, if not and error must be returned.
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
