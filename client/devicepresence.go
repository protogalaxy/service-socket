package client

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/arjantop/cuirass"
	"github.com/arjantop/saola/httpservice"
	"github.com/golang/glog"
	"github.com/protogalaxy/service-socket/socket"
	"golang.org/x/net/context"
)

type Status string

const (
	Online  Status = "online"
	Offline Status = "offline"
)

type DevicePresence interface {
	SetDeviceStatus(ctx context.Context, deviceId socket.ID, userID string, status Status) error
}

var _ DevicePresence = (*DevicePresenceClient)(nil)

type DevicePresenceClient struct {
	Client   *httpservice.Client
	Executor cuirass.Executor
}

func (c *DevicePresenceClient) SetDeviceStatus(ctx context.Context, deviceID socket.ID, userID string, status Status) error {
	cmd := cuirass.NewCommand("SetDeviceStatus", func(ctx context.Context) (interface{}, error) {
		requestJson := fmt.Sprintf(`{"user_id": "%s", "status": "%s"}`, userID, status)
		url := fmt.Sprintf("http://localhost:10000/status/websocket/%s", deviceID)
		req, err := http.NewRequest("PUT", url, strings.NewReader(requestJson))
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			glog.Fatal("Problem creating request: ", err)
		}

		res, err := c.Client.Do(ctx, req)
		if res.StatusCode == http.StatusOK {
			glog.V(3).Infof("Set device status to '%s'", status)
			return nil, nil
		} else {
			glog.Error("Setting device status failed")
			return nil, errors.New("TODO")
		}
	}).Build()

	_, err := c.Executor.Exec(ctx, cmd)
	return err
}
