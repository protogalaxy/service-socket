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
package service

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/arjantop/saola/httpservice"
	"github.com/golang/glog"
	"github.com/protogalaxy/common/serviceerror"
	"github.com/protogalaxy/service-socket/socket"
	"golang.org/x/net/context"
)

// SocketSendMsg is a service that provides sending messages to registered sockets.
type SocketSendMsg struct {
	Sockets socket.Registry
}

func decodeMessage(ps httpservice.Params, v *socket.Message, body io.Reader) error {
	deviceID, err := socket.ParseID(ps.Get("deviceID"))
	if err != nil {
		eres := serviceerror.BadRequest("invalid_request", "Invalid device id")
		eres.Cause = err
		return eres
	}

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return serviceerror.InternalServerError("server_error", "Unable to read request body", err)
	}

	v.SocketID = deviceID
	v.Data = data
	return nil
}

// DoHTTP implements saola.HttpService.
func (h *SocketSendMsg) DoHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var msg socket.Message
	err := decodeMessage(httpservice.GetParams(ctx), &msg, r.Body)
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte("{}\n"))
	return nil
}

// Do implements saola.Service.
func (h *SocketSendMsg) Do(ctx context.Context) error {
	return httpservice.Do(h, ctx)
}

// Name implements saola.Service.
func (h *SocketSendMsg) Name() string {
	return "websocketsendmessage"
}
