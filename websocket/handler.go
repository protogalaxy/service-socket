package websocket

import (
	"github.com/golang/glog"
	"github.com/protogalaxy/service-socket/client"
	"github.com/protogalaxy/service-socket/socket"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
)

type ConnectionHandler struct {
	Registry             *socket.RegistryServer
	DevicePresenceClient client.DevicePresence
}

type MsgConn struct {
	*websocket.Conn
}

func (c *MsgConn) ReadMessage() ([]byte, error) {
	var data []byte
	err := websocket.Message.Receive(c.Conn, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (h *ConnectionHandler) Handler() websocket.Handler {
	return websocket.Handler(func(raw *websocket.Conn) {
		ws := &MsgConn{raw}
		defer ws.Close()

		writer := socket.NewMessageWriter(ws)
		reader := socket.NewMessageReader(ws)
		writer.Reader = reader
		reader.Writer = writer

		socketID, err := h.Registry.Register(writer.Messages())
		if err != nil {
			glog.Errorf("Could not register socket: %s", err)
		}
		glog.V(2).Infof("New websocket connection %s", socketID)

		defer func() {
			done := make(chan struct{})
			go writer.DiscardUntil(done)
			h.Registry.Unregister(socketID)
			close(done)
		}()

		userID := "user1"
		glog.V(2).Infof("Athenticated user %s on websocket connection %s", userID, socketID)

		ctx := context.Background()
		err = h.DevicePresenceClient.SetDeviceStatus(ctx, socketID, userID, client.Online)
		if err != nil {
			glog.Errorf("Problem setting device status: %s", err)
			return
		}

		go writer.Run()
		go func() {
			for {
				// TODO: Don't discard all messages
				<-reader.Messages()
			}
		}()
		reader.Run()
	})
}
