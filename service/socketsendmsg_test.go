package service_test

import (
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
	json := `{"data":"aGVsbG8="}`
	ps := httpservice.EmptyParams()
	ps.Set("deviceId", "123")
	r, _ := http.NewRequest("POST", "", strings.NewReader(json))
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

func TestSocketSendMsgInvalidDeviceId(t *testing.T) {
	t.Parallel()
	s := service.SocketSendMsg{
		Sockets: nil,
	}
	ps := httpservice.EmptyParams()
	ps.Set("deviceId", "abc")
	r, _ := http.NewRequest("POST", "", nil)
	err := s.DoHTTP(httpservice.WithParams(context.Background(), ps), nil, r)
	if err == nil {
		t.Fatal("Expecting error but got none")
	}

	ps.Set("deviceId", "")
	r, _ = http.NewRequest("POST", "", nil)
	err = s.DoHTTP(httpservice.WithParams(context.Background(), ps), nil, r)
	if err == nil {
		t.Fatal("Expecting error but got none")
	}
}

func TestSocketSendMsgInvalidJsonBody(t *testing.T) {
	t.Parallel()
	s := service.SocketSendMsg{
		Sockets: nil,
	}
	json := `{"data:"aGVsbG8="}`
	ps := httpservice.EmptyParams()
	ps.Set("deviceId", "123")
	r, _ := http.NewRequest("POST", "", strings.NewReader(json))
	err := s.DoHTTP(httpservice.WithParams(context.Background(), ps), nil, r)
	if err == nil {
		t.Fatal("Expecting error but got none")
	}
}

func TestSocketSendMsgInvalidDataEncoding(t *testing.T) {
	t.Parallel()
	s := service.SocketSendMsg{
		Sockets: nil,
	}
	json := `{"data":"aGVsbG8"}`
	ps := httpservice.EmptyParams()
	ps.Set("deviceId", "123")
	r, _ := http.NewRequest("POST", "", strings.NewReader(json))
	err := s.DoHTTP(httpservice.WithParams(context.Background(), ps), nil, r)
	if err == nil {
		t.Fatal("Expecting error but got none")
	}
}
