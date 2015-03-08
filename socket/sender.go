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

//go:generate protoc --go_out=plugins=grpc:. -I ../protos ../protos/socket.proto

package socket

import (
	"errors"

	"github.com/golang/glog"
	"golang.org/x/net/context"
)

type Sender struct {
	Sockets Registry
}

func validateRequest(req *SendRequest) error {
	if len(req.Data) == 0 {
		return errors.New("empty message")
	}
	return nil
}

func (s *Sender) SendMessage(ctx context.Context, req *SendRequest) (*SendReply, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	msg := Message{
		SocketID: ID(req.SocketId),
		Data:     req.Data,
	}

	select {
	case s.Sockets.Messages() <- msg:
		glog.V(3).Info("Message sent")
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return &SendReply{}, nil
}
