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
