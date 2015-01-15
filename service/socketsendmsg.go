package service

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/arjantop/saola/httpservice"
	"github.com/golang/glog"
	"github.com/protogalaxy/common/serviceerror"
	"github.com/protogalaxy/service-socket/socket"
	"golang.org/x/net/context"
)

// SeocketSendMsg is a service that provides sending messages to registered sockets.
type SocketSendMsg struct {
	Sockets *socket.SocketRegistry
}

type Message struct {
	EncodedData string `json:"data"`
}

func decodeMessage(ps httpservice.Params, body io.ReadCloser) (socket.Message, error) {
	deviceId, err := strconv.ParseInt(ps.Get("deviceId"), 10, 64)
	if err != nil {
		return socket.Message{}, serviceerror.BadRequest("invalid device id", err)
	}

	decoder := json.NewDecoder(body)
	var message Message
	err = decoder.Decode(&message)
	if err != nil {
		return socket.Message{}, serviceerror.BadRequest("problem decoding request body", err)
	}

	data, err := base64.StdEncoding.DecodeString(message.EncodedData)
	if err != nil {
		return socket.Message{}, serviceerror.BadRequest("invalid base64 data encoding", err)
	}

	return socket.Message{
		SocketId: socket.Id(deviceId),
		Data:     data,
	}, nil
}

// DoHTTP implements saola.HttpService.
func (h *SocketSendMsg) DoHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	msg, err := decodeMessage(httpservice.GetParams(ctx), r.Body)
	if err != nil {
		return err
	}
	select {
	case h.Sockets.Messages() <- msg:
		glog.V(3).Info("Message sent")
	case <-ctx.Done():
		return ctx.Err()
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json; utf-8")
	w.Write([]byte("{}\n"))
	return nil
}

// Do implements saola.Service.
func (h *SocketSendMsg) Do(ctx context.Context) error {
	return httpservice.Do(h, ctx)
}

func (h *SocketSendMsg) Name() string {
	return "websocketsendmessage"
}
