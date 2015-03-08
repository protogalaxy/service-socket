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
package main

import (
	"flag"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/arjantop/cuirass"
	"github.com/arjantop/saola/httpservice"
	"github.com/arjantop/vaquita"
	"github.com/golang/glog"
	"github.com/protogalaxy/service-socket/client"
	"github.com/protogalaxy/service-socket/socket"
	"github.com/protogalaxy/service-socket/websocket"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	config := vaquita.NewEmptyMapConfig()
	exec := cuirass.NewExecutor(config)

	socketRegistry := socket.NewRegistry()
	go socketRegistry.Run()

	httpClient := &httpservice.Client{
		Transport: &http.Transport{},
	}
	devicePresence := &client.DevicePresenceClient{
		Client:   httpClient,
		Executor: exec,
	}

	connHandler := websocket.ConnectionHandler{
		Registry:             socketRegistry,
		DevicePresenceClient: devicePresence,
	}
	go func() {
		http.Handle("/", connHandler.Handler())
		glog.Fatal(http.ListenAndServe(":8080", nil))
	}()

	s, err := net.Listen("tcp", ":9090")
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	socket.RegisterSenderServer(grpcServer, &socket.Sender{
		Sockets: socketRegistry,
	})
	grpcServer.Serve(s)
}
