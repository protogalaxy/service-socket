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
	"io"
	"net/http"

	"github.com/protogalaxy/service-socket/Godeps/_workspace/src/github.com/golang/glog"
	"github.com/protogalaxy/service-socket/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/protogalaxy/service-socket/devicepresence"
	"github.com/protogalaxy/service-socket/messagebroker"
	"github.com/protogalaxy/service-socket/socket"
)

type StateFunc func(*States) *StateFunc

func Run(s *States) {
	curr := s.Initial()
	for curr != nil {
		curr = (*curr)(s)
	}
}

type States struct {
	Registry       socket.Registry
	DevicePresence devicepresence.PresenceManagerClient
	MessageBroker  messagebroker.BrokerClient
	Conn           Conn
	Messages       chan []byte
	socketID       socket.ID
	userID         string
}

type Conn interface {
	Request() *http.Request
	ReadMessage() ([]byte, error)
	io.Writer
}

var (
	AuthenticateUser StateFunc = (*States).authenticateUser
	RegisterSocket   StateFunc = (*States).registerSocket
	SetDeviceStatus  StateFunc = (*States).setDeviceStatus
	HandleMessages   StateFunc = (*States).handleMessages
)

func (s *States) Initial() *StateFunc {
	return &AuthenticateUser
}

func (s *States) authenticateUser() *StateFunc {
	c, err := s.Conn.Request().Cookie("auth")
	if err != nil {
		glog.Info("Missing authentication cookie")
		return nil
	}
	// TODO: call the auth service
	s.userID = c.Value
	glog.V(2).Infof("Athenticated as user %s", s.userID)
	return &RegisterSocket
}

func (s *States) registerSocket() *StateFunc {
	socketID, err := s.Registry.Register(s.Messages)
	if err != nil {
		glog.Errorf("Could not register socket: %s", err)
		return nil
	}
	s.socketID = socketID
	return &SetDeviceStatus
}

func (s *States) setDeviceStatus() *StateFunc {
	// TODO: add timeout
	ctx := context.Background()
	_, err := s.DevicePresence.SetStatus(ctx, &devicepresence.StatusRequest{
		Device: &devicepresence.Device{
			Id:     s.socketID.String(),
			Type:   devicepresence.Device_WS,
			UserId: s.userID,
			Status: devicepresence.Device_ONLINE,
		},
	})
	if err != nil {
		glog.Errorf("Problem setting device status: %s", err)
		// TODO: transition to disconnected state
		return nil
	}
	return &HandleMessages
}

func (s *States) handleMessages() *StateFunc {
	writer := socket.NewMessageWriter(s.Conn, s.Messages)
	reader := socket.NewMessageReader(s.Conn)
	writer.Reader = reader
	reader.Writer = writer

	go writer.Run()
	go func() {
		for {
			// TODO: timeout
			ctx := context.Background()
			msg := <-reader.Messages()
			_, err := s.MessageBroker.Route(ctx, &messagebroker.RouteRequest{
				Data: msg,
			})
			if err != nil {
				glog.Errorf("handling message: %s", err)
			}
		}
	}()
	reader.Run()

	return nil
}
