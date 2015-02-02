package httpservice

import (
	"net/http"

	"github.com/arjantop/saola"
	"golang.org/x/net/context"
)

type CancellableRoundTripper interface {
	http.RoundTripper
	CancelRequest(*http.Request)
}

type ClientRequest struct {
	Request  *http.Request
	Response *http.Response
}

type clientRequest struct{}

func withClientRequest(ctx context.Context, cr *ClientRequest) context.Context {
	return context.WithValue(ctx, clientRequest{}, cr)
}

func GetClientRequest(ctx context.Context) *ClientRequest {
	if r, ok := ctx.Value(clientRequest{}).(*ClientRequest); ok {
		return r
	}
	panic("missing client request")
}

type Client struct {
	Filter    saola.Filter
	service   saola.Service
	Transport CancellableRoundTripper
}

type result struct {
	Response *http.Response
	Error    error
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if c.service == nil {
		s := newClientService(c.Transport)
		if c.Filter != nil {
			c.service = saola.Apply(s, c.Filter)
		} else {
			c.service = s
		}
	}
	cr := &ClientRequest{Request: req}
	err := c.service.Do(withClientRequest(ctx, cr))
	return cr.Response, err
}

func newClientService(tr CancellableRoundTripper) saola.Service {
	client := http.Client{Transport: tr}
	return saola.FuncService(func(ctx context.Context) error {
		cr := GetClientRequest(ctx)
		r := make(chan result, 1)
		go func() {
			resp, err := client.Do(cr.Request)
			r <- result{resp, err}
		}()
		select {
		case <-ctx.Done():
			tr.CancelRequest(cr.Request)
			<-r
			return ctx.Err()
		case result := <-r:
			cr.Response = result.Response
			return result.Error
		}
	})
}
