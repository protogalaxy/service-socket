package httpservice_test

import (
	"net/http"
	"testing"

	"github.com/arjantop/saola"
	"github.com/arjantop/saola/httpservice"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestRequestLogFilter(t *testing.T) {
	w, _ := newTestingResponseWriter()
	req, _ := http.NewRequest("PUT", "http://localhost:8080/foo", nil)
	ctx := httpservice.WithServerRequest(context.Background(), w, req)
	var logEntry httpservice.LogEntry
	s := saola.Apply(saola.NoopService{}, httpservice.NewRequestLogFilter(func(e httpservice.LogEntry) {
		logEntry = e
	}))
	assert.NoError(t, s.Do(ctx))
	assert.Equal(t, "PUT", logEntry.RequestMethod)
	assert.Equal(t, "/foo", logEntry.RequestPath)
	assert.Equal(t, 200, logEntry.StatusCode)
	assert.True(t, logEntry.Latency > 0)
}

func TestRequestLogFilterNoStatusCodeInterceptor(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/bar", nil)
	ctx := httpservice.WithServerRequest(context.Background(), NoopResponseWriter{}, req)
	var logEntry httpservice.LogEntry
	s := saola.Apply(saola.NoopService{}, httpservice.NewRequestLogFilter(func(e httpservice.LogEntry) {
		logEntry = e
	}))
	assert.NoError(t, s.Do(ctx))
	assert.Equal(t, 0, logEntry.StatusCode, "No interceptor present")
}

func BenchmarkRequestLog(b *testing.B) {
	req, _ := http.NewRequest("POST", "http://localhost:8080/foo", nil)
	ctx := httpservice.WithServerRequest(context.Background(), NoopResponseWriter{}, req)
	s := saola.Apply(saola.NoopService{}, httpservice.NewRequestLogFilter(func(e httpservice.LogEntry) {}))
	for i := 0; i < b.N; i++ {
		s.Do(ctx)
	}
}
