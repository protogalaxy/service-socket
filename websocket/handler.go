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
package websocket

import (
	"github.com/protogalaxy/service-socket/Godeps/_workspace/src/golang.org/x/net/websocket"
	"github.com/protogalaxy/service-socket/devicepresence"
	"github.com/protogalaxy/service-socket/messagebroker"
	"github.com/protogalaxy/service-socket/socket"
)

type ConnectionHandler struct {
	Registry       *socket.RegistryServer
	DevicePresence devicepresence.PresenceManagerClient
	MessageBroker  messagebroker.BrokerClient
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
		s := States{
			Registry:       h.Registry,
			DevicePresence: h.DevicePresence,
			MessageBroker:  h.MessageBroker,
			Conn:           ws,
			Messages:       make(chan []byte, 10),
		}

		// TODO: set device status to offline
		defer func() {
			if s.socketID != 0 {
				h.Registry.Unregister(s.socketID)
			}
		}()

		Run(&s)
	})
}
