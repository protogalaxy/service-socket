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

// MessageWriter is a worker that writes the messages to the specified io.Writer.
// If the Reader is set it will be closed after the Run terminates.
// The Reader should not be set while the writer is running.
type MessageWriter struct {
	w         io.Writer
	messages  chan []byte
	close     chan struct{}
	closeOnce sync.Once
	Reader    io.Closer
}

// NewMessageWriter constructs a new MessageWriter for a given writer.
func NewMessageWriter(w io.Writer) *MessageWriter {
	return &MessageWriter{
		w:        w,
		messages: make(chan []byte),
		close:    make(chan struct{}),
	}
}

// Run reads messages from its message channel and writes them to the writer.
// Method terminates if a write error occurs or the writer is explicitly closed.
// If set the Reader is closed before returning.
func (w *MessageWriter) Run() {
	defer func() {
		if w.Reader != nil {
			w.Reader.Close()
		}
	}()
	for {
		select {
		case msg := <-w.messages:
			glog.V(4).Info("Writing message")
			_, err := w.w.Write(msg)
			if err != nil {
				glog.Warning("Unable to write: ", err)
				return
			}
		case <-w.close:
			glog.Info("Closing message writer")
			return
		}
	}
}

// DiscardUntil closes the writer and then consumes the messages from writer's message
// channel until provided done channel is not closed.
func (w *MessageWriter) DiscardUntil(done chan struct{}) {
	w.Close()
	for {
		select {
		case <-w.messages:
			glog.V(4).Infof("Discarding message")
		case <-done:
			return
		}
	}
}

// Messages returns a send only channel of messages that will be written.
func (w *MessageWriter) Messages() chan<- []byte {
	return w.messages
}

// Close implements a Closer interface.
// The method is idempotents and can be called multiple times.
func (w *MessageWriter) Close() error {
	w.closeOnce.Do(func() { close(w.close) })
	return nil
}
