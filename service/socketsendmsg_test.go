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
package service_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arjantop/saola/httpservice"
	"github.com/protogalaxy/service-socket/service"
	"github.com/protogalaxy/service-socket/socket"
	"golang.org/x/net/context"
)

func TestSocketSendMsgMessageIsSent(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	reg := socket.NewRegistry()
	go reg.Run()
	defer reg.Close()
	s := service.SocketSendMsg{
		Sockets: reg,
	}
	ps := httpservice.EmptyParams()
	ps.Set("deviceID", "123")
	r, _ := http.NewRequest("POST", "", strings.NewReader("abc"))
	err := s.DoHTTP(httpservice.WithParams(context.Background(), ps), w, r)
	if err != nil {
		t.Fatalf("Expecting no error but got: %s", err)
	}
	if w.Code != http.StatusAccepted {
		t.Fatalf("Unexpected status code: %d != %d", http.StatusAccepted, w.Code)
	}
	if strings.TrimSpace(w.Body.String()) != "{}" {
		t.Fatalf("Body should be an empty json document but got '%s'", w.Body.String())
	}
}

func TestSocketSendMsgInvalidDeviceID(t *testing.T) {
	t.Parallel()
	s := service.SocketSendMsg{
		Sockets: nil,
	}
	ps := httpservice.EmptyParams()
	ps.Set("deviceID", "z")
	r, _ := http.NewRequest("POST", "", nil)
	err := s.DoHTTP(httpservice.WithParams(context.Background(), ps), nil, r)
	if err == nil {
		t.Fatal("Expecting error but got none")
	}

	ps.Set("deviceID", "")
	r, _ = http.NewRequest("POST", "", nil)
	err = s.DoHTTP(httpservice.WithParams(context.Background(), ps), nil, r)
	if err == nil {
		t.Fatal("Expecting error but got none")
	}
}

type errorReader struct{}

func (r errorReader) Read(b []byte) (int, error) {
	return 0, errors.New("error")
}

func TestSocketSendMsgBodyReadError(t *testing.T) {
	t.Parallel()
	s := service.SocketSendMsg{
		Sockets: nil,
	}
	ps := httpservice.EmptyParams()
	ps.Set("deviceID", "123")
	r, _ := http.NewRequest("POST", "", errorReader{})
	err := s.DoHTTP(httpservice.WithParams(context.Background(), ps), nil, r)
	if err == nil {
		t.Fatal("Expecting error but got none")
	}
}
