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
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/arjantop/cuirass"
	"github.com/arjantop/saola"
	"github.com/arjantop/saola/httpservice"
	"github.com/arjantop/vaquita"
	"github.com/golang/glog"
	"github.com/protogalaxy/common/serviceerror"
	"github.com/protogalaxy/service-socket/client"
	"github.com/protogalaxy/service-socket/service"
	"github.com/protogalaxy/service-socket/socket"
	"github.com/protogalaxy/service-socket/websocket"
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
		glog.Fatal(http.ListenAndServe(":10300", nil))
	}()

	endpoint := httpservice.NewEndpoint()

	endpoint.POST("/websocket/:deviceID/send", saola.Apply(
		&service.SocketSendMsg{
			Sockets: socketRegistry,
		},
		httpservice.NewCancellationFilter(),
		serviceerror.NewErrorResponseFilter(),
		serviceerror.NewErrorLoggerFilter()))

	log.Fatal(httpservice.Serve(":10301", saola.Apply(
		endpoint,
		httpservice.NewStdRequestLogFilter())))
}
