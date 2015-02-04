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

	"github.com/protogalaxy/service-socket/client"
	"github.com/protogalaxy/service-socket/socket"
	"golang.org/x/net/context"
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
	OnSetDeviceStatus func(ctx context.Context, deviceId socket.ID, userID string, status client.Status) error
}

func (m *DevicePresenceMock) SetDeviceStatus(ctx context.Context, deviceID socket.ID, userID string, status client.Status) error {
	return m.OnSetDeviceStatus(ctx, deviceID, userID, status)
}

func TestStatesSetDeviceStatus(t *testing.T) {
	s := &States{
		DevicePresenceClient: &DevicePresenceMock{
			OnSetDeviceStatus: func(ctx context.Context, deviceID socket.ID, userID string, status client.Status) error {
				if deviceID != socket.ID(9) {
					t.Errorf("Unexpected device id: %s", deviceID)
				}
				if userID != "13" {
					t.Errorf("Unexpected user id: %s", userID)
				}
				if status != client.Online {
					t.Errorf("Unexpected device status: %s", status)
				}
				return nil
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
		DevicePresenceClient: &DevicePresenceMock{
			OnSetDeviceStatus: func(ctx context.Context, deviceID socket.ID, userID string, status client.Status) error {
				return errors.New("error")
			},
		},
	}
	next := s.setDeviceStatus()

	if next != nil {
		t.Errorf("Invalid next state")
	}
}
