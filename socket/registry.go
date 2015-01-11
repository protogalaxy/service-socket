package socket

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/golang/glog"
)

// Id identifies the registered socket in the registry.
type Id int64

// SocketRegistry allows sockets to register themselves with a channel they want to receive on
// and all the received messages will be routed to the correct socket's channel.
type SocketRegistry struct {
	activeSockets map[Id]chan<- []byte
	done          chan struct{}
	closeOnce     sync.Once
	messages      chan Message
	register      chan registerSocket
	unregister    chan Id
}

// NewRegistry constructs a new socket registry that is ready to be run.
func NewRegistry() *SocketRegistry {
	return &SocketRegistry{
		activeSockets: make(map[Id]chan<- []byte),
		done:          make(chan struct{}),
		messages:      make(chan Message),
		register:      make(chan registerSocket),
		unregister:    make(chan Id),
	}
}

// Run starts to handle received events and executes appropriate actions.
// The method blocks until the registry is closed and is normally be run in
// a goroutine.
func (r *SocketRegistry) Run() {
	for {
		select {
		case <-r.done:
			glog.Info("Shutting down socket registry")
			return
		case m := <-r.messages:
			if c, ok := r.activeSockets[m.SocketId]; ok {
				c <- m.Data
			} else {
				glog.Warningf("Socket not found: %s", m.SocketId)
			}
		case m := <-r.register:
			glog.Infof("Registering socket id=%d", m.SocketId)
			r.activeSockets[m.SocketId] = m.Messages
		case socketId := <-r.unregister:
			glog.Infof("Unregistering socket id=%d", socketId)
			delete(r.activeSockets, socketId)
		}
	}
}

// Close implements the Closer interface.
// The receiving loop terminates (if running), messages channel is not closed.
func (r *SocketRegistry) Close() error {
	r.closeOnce.Do(func() { close(r.done) })
	return nil
}

// Message is a single message that will be routed to the receiving socket channel
// with the matching id.
type Message struct {
	SocketId Id
	Data     []byte
}

// Messages returns a send-only channel that receives the messages that are then
// routed to the matching sockets.
func (r *SocketRegistry) Messages() chan<- Message {
	return r.messages
}

// registerSocket represents information needed for registering a socket's channel.
type registerSocket struct {
	SocketId Id
	Messages chan<- []byte
}

// Register registers a channel that messages are routed to for the socket.
// A new socket id is generated and returned that can be used to unregister
// the channel for receiving later.
// Error can occur if the random id could not be generated.
func (r *SocketRegistry) Register(messages chan<- []byte) (Id, error) {
	socketIdBig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return 0, fmt.Errorf("generating socket id: %s", err)
	}
	socketId := Id(socketIdBig.Int64())

	r.register <- registerSocket{
		SocketId: socketId,
		Messages: messages,
	}
	return socketId, nil
}

// Unregister unregisters a receiving channel from the registry.
// Once unregistered no more messages are going to be received.
// Until successfully unregistered the client should keep receiving the messages
// on the registered channel.
func (r *SocketRegistry) Unregister(socketId Id) {
	r.unregister <- socketId
}
