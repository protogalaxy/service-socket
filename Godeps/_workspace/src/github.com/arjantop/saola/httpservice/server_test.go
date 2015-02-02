package httpservice_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arjantop/saola/httpservice"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestServerParamsInContext(t *testing.T) {
	params := httpservice.EmptyParams()
	params.Set("k", "v")

	retrieved := httpservice.GetParams(httpservice.WithParams(context.Background(), params))
	assert.Equal(t, "v", retrieved.Get("k"))
}

func TestServerParamsNotInContext(t *testing.T) {
	retrieved := httpservice.GetParams(context.Background())
	assert.Equal(t, "", retrieved.Get("k"))
}

func TestServerParamsMultipleSet(t *testing.T) {
	params := httpservice.EmptyParams()
	params.Set("k", "v")
	params.Set("k", "v2")

	assert.Equal(t, "v2", params.Get("k"))
}

func TestServerEndpointGET(t *testing.T) {
	endpoint := httpservice.NewEndpoint()
	endpoint.GET("/hello/:name", httpservice.FuncService(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		params := httpservice.GetParams(ctx)
		fmt.Fprintf(w, params.Get("name"))
		return nil
	}))

	req, err := http.NewRequest("GET", "http://example.com/hello/bob", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	endpoint.DoHTTP(context.Background(), w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "bob", w.Body.String())
}

func TestServerEndpointPOST(t *testing.T) {
	endpoint := httpservice.NewEndpoint()
	endpoint.POST("/hello/:name", httpservice.FuncService(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		params := httpservice.GetParams(ctx)
		fmt.Fprintf(w, params.Get("name"))
		return nil
	}))

	req, err := http.NewRequest("POST", "http://example.com/hello/lucian", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	endpoint.DoHTTP(context.Background(), w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "lucian", w.Body.String())
}

func TestServerEndpointPUT(t *testing.T) {
	endpoint := httpservice.NewEndpoint()
	endpoint.PUT("/hello/:name", httpservice.FuncService(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		params := httpservice.GetParams(ctx)
		fmt.Fprintf(w, params.Get("name"))
		return nil
	}))

	req, err := http.NewRequest("PUT", "http://example.com/hello/john", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	endpoint.DoHTTP(context.Background(), w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "john", w.Body.String())
}

func TestServerEndpointService(t *testing.T) {
	endpoint := httpservice.NewEndpoint()
	endpoint.PUT("/hello/:name", httpservice.FuncService(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		fmt.Fprintf(w, "response")
		return nil
	}))

	req, err := http.NewRequest("PUT", "http://example.com/hello/john", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	ctx := httpservice.WithServerRequest(context.Background(), w, req)
	endpoint.Do(ctx)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "response", w.Body.String())
}
