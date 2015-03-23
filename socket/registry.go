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
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"sync"

	"github.com/protogalaxy/service-socket/Godeps/_workspace/src/github.com/golang/glog"
)

// ID identifies the registered socket in the registry.
type ID int64

func (i ID) String() string {
	return fmt.Sprintf("%x", int64(i))
}

func ParseID(v string) (ID, error) {
	id, err := strconv.ParseInt(v, 16, 64)
	return ID(id), err
}

// Registry allows sockets to register themselves with a channel they want to receive on
// and all the received messages will be routed to the correct socket's channel.
type Registry interface {
	// Messages returns a send-only channel that receives the messages that are then
	// routed to the matching sockets.
	Messages() chan<- Message

	// Register registers a channel that messages are routed to for the socket.
	// A new socket id is generated and returned that can be used to unregister
	// the channel.
	Register(messages chan<- []byte) (ID, error)

	// Unregister unregisters a receiving channel from the registry.
	// Once unregistered no more messages are going to be received.
	// Until successfully unregistered the client should keep receiving the messages
	// on the registered channel.
	Unregister(socketId ID)
}

var _ Registry = (*RegistryServer)(nil)

// RegistryServer is an implementation of Registry using an event loop to handle the
// received messages.
type RegistryServer struct {
	activeSockets map[ID]chan<- []byte

	closeOnce sync.Once
	done      chan struct{}

	messages   chan Message
	register   chan registerSocket
	unregister chan ID
}

// NewRegistry constructs a new socket registry that is ready to be run.
func NewRegistry() *RegistryServer {
	return &RegistryServer{
		activeSockets: make(map[ID]chan<- []byte),
		done:          make(chan struct{}),
		messages:      make(chan Message),
		register:      make(chan registerSocket),
		unregister:    make(chan ID),
	}
}

// Run starts to handle received events and executes appropriate actions.
// The method blocks until the registry is closed and is normally be run in
// a goroutine.
func (r *RegistryServer) Run() {
	for {
		select {
		case <-r.done:
			glog.Info("Shutting down socket registry")
			return
		case m := <-r.messages:
			if c, ok := r.activeSockets[m.SocketID]; ok {
				select {
				case c <- m.Data:
					glog.V(4).Info("Message sent to socket")
				default:
					glog.Warning("Socket queue full")
				}
			} else {
				glog.Warningf("Socket not found: %s", m.SocketID)
			}
		case m := <-r.register:
			glog.Infof("Registering socket: %s", m.SocketId)
			r.activeSockets[m.SocketId] = m.Messages
		case socketId := <-r.unregister:
			glog.Infof("Unregistering socket: %s", socketId)
			delete(r.activeSockets, socketId)
		}
	}
}

// Close implements the Closer interface.
// The receiving loop terminates (if running), messages channel is not closed.
func (r *RegistryServer) Close() error {
	r.closeOnce.Do(func() { close(r.done) })
	return nil
}

// Message is a single message that will be routed to the receiving socket channel
// with the matching id.
type Message struct {
	SocketID ID
	Data     []byte
}

// Messages implements the Registry interface.
func (r *RegistryServer) Messages() chan<- Message {
	return r.messages
}

// registerSocket represents information needed for registering a socket's channel.
type registerSocket struct {
	SocketId ID
	Messages chan<- []byte
}

// Register implements the Registry interface.
// Error can occur if the random id could not be generated.
func (r *RegistryServer) Register(messages chan<- []byte) (ID, error) {
	socketIdBig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return 0, fmt.Errorf("generating socket id: %s", err)
	}
	socketId := ID(socketIdBig.Int64())

	r.register <- registerSocket{
		SocketId: socketId,
		Messages: messages,
	}
	return socketId, nil
}

// Unregister implements the Registry interface.
func (r *RegistryServer) Unregister(socketId ID) {
	r.unregister <- socketId
}
