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
	"errors"
	"net/http"
	"testing"

	"github.com/protogalaxy/service-socket/devicepresence"
	"github.com/protogalaxy/service-socket/socket"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type ConnMock struct {
	OnRequest     func() *http.Request
	OnReadMessage func() ([]byte, error)
	OnWrite       func([]byte) (int, error)
}

func (m *ConnMock) Request() *http.Request {
	return m.OnRequest()
}

func (m *ConnMock) ReadMessage() ([]byte, error) {
	return m.OnReadMessage()
}

func (m *ConnMock) Write(p []byte) (int, error) {
	return m.OnWrite(p)
}

func TestStatesAuthenticateUser(t *testing.T) {
	s := &States{
		Conn: &ConnMock{
			OnRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "", nil)
				req.AddCookie(&http.Cookie{
					Name:  "auth",
					Value: "user1",
				})
				return req
			},
		},
	}
	next := s.authenticateUser()
	if next != &RegisterSocket {
		t.Errorf("Invalid next state")
	}
	if s.userID != "user1" {
		t.Errorf("Unexpected user id: %s", s.userID)
	}
}

func TestStatesAuthenticateUserMissingCookie(t *testing.T) {
	s := &States{
		Conn: &ConnMock{
			OnRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "", nil)
				return req
			},
		},
	}
	next := s.authenticateUser()
	if next != nil {
		t.Errorf("Invalid next state")
	}
}

type RegistryMock struct {
	OnMessages   func() chan<- socket.Message
	OnRegister   func(messages chan<- []byte) (socket.ID, error)
	OnUnregister func(socketID socket.ID)
}

func (m *RegistryMock) Messages() chan<- socket.Message {
	return m.OnMessages()
}

func (m *RegistryMock) Register(messages chan<- []byte) (socket.ID, error) {
	return m.OnRegister(messages)
}

func (m *RegistryMock) Unregister(socketID socket.ID) {
	m.OnUnregister(socketID)
}

func TestStatesRegisterSocket(t *testing.T) {
	s := &States{
		Registry: &RegistryMock{
			OnRegister: func(msgs chan<- []byte) (socket.ID, error) {
				return 123, nil
			},
		},
	}
	next := s.registerSocket()
	if next != &SetDeviceStatus {
		t.Errorf("Invalid next state")
	}
	if s.socketID != 123 {
		t.Errorf("Unexpected socket id: %d", s.socketID)
	}
}

func TestStatesRegisterSocketError(t *testing.T) {
	s := &States{
		Registry: &RegistryMock{
			OnRegister: func(msgs chan<- []byte) (socket.ID, error) {
				return 0, errors.New("error")
			},
		},
	}
	next := s.registerSocket()
	if next != nil {
		t.Errorf("Invalid next state")
	}
}

type DevicePresenceMock struct {
	OnSetStatus func(context.Context, *devicepresence.StatusRequest) (*devicepresence.StatusReply, error)
}

func (m *DevicePresenceMock) SetStatus(ctx context.Context, req *devicepresence.StatusRequest, opts ...grpc.CallOption) (*devicepresence.StatusReply, error) {
	return m.OnSetStatus(ctx, req)
}

func TestStatesSetDeviceStatus(t *testing.T) {
	s := &States{
		DevicePresence: &DevicePresenceMock{
			OnSetStatus: func(ctx context.Context, req *devicepresence.StatusRequest) (*devicepresence.StatusReply, error) {
				if req.Device == nil {
					t.Errorf("Missing device")
				}
				expected := devicepresence.Device{
					Id:     "9",
					Type:   devicepresence.Device_WS,
					UserId: "13",
					Status: devicepresence.Device_ONLINE,
				}
				if expected != *req.Device {
					t.Errorf("Unexpected device: %#v != %#v", expected, req.Device)
				}
				return &devicepresence.StatusReply{}, nil
			},
		},
	}
	s.socketID = 9
	s.userID = "13"
	next := s.setDeviceStatus()

	if next != &HandleMessages {
		t.Errorf("Invalid next state")
	}
}

func TestStatesSetDeviceStatusError(t *testing.T) {
	s := &States{
		DevicePresence: &DevicePresenceMock{
			OnSetStatus: func(ctx context.Context, req *devicepresence.StatusRequest) (*devicepresence.StatusReply, error) {
				return nil, errors.New("error")
			},
		},
	}
	next := s.setDeviceStatus()

	if next != nil {
		t.Errorf("Invalid next state")
	}
}
