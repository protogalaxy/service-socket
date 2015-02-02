package httpservice_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/arjantop/saola"
	"github.com/arjantop/saola/httpservice"
	"github.com/arjantop/saola/stats/statstest"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func newContext(method string) context.Context {
	w, _ := newTestingResponseWriter()
	req, _ := http.NewRequest(method, "http://localhost:8080/foo", nil)
	return httpservice.WithServerRequest(context.Background(), w, req)
}

func TestResponseStatsFilter(t *testing.T) {
	r := statstest.NewRecorder()
	f := httpservice.NewResponseStatsFilter(r)
	err := f.Do(newContext("GET"), saola.FuncService(func(ctx context.Context) error {
		return nil
	}))
	assert.NoError(t, err)
	assert.Equal(t, 1, r.CounterValue("func.http.status.200"))
	assert.Equal(t, 1, r.CounterValue("func.http.status.2xx"))
	assert.True(t, r.TimerValue("func.http.time.200") > 0)
	assert.True(t, r.TimerValue("func.http.time.2xx") > 0)

	err1 := f.Do(newContext("GET"), saola.FuncService(func(ctx context.Context) error {
		r := httpservice.GetServerRequest(ctx)
		r.Writer.WriteHeader(http.StatusInternalServerError)
		return errors.New("error")
	}))

	err2 := f.Do(newContext("GET"), saola.FuncService(func(ctx context.Context) error {
		r := httpservice.GetServerRequest(ctx)
		r.Writer.WriteHeader(http.StatusNotFound)
		return errors.New("error")
	}))

	assert.Error(t, err1)
	assert.Equal(t, 1, r.CounterValue("func.http.status.500"))
	assert.Equal(t, 1, r.CounterValue("func.http.status.5xx"))
	assert.True(t, r.TimerValue("func.http.time.500") > 0)
	assert.True(t, r.TimerValue("func.http.time.5xx") > 0)

	assert.Error(t, err2)
	assert.Equal(t, 1, r.CounterValue("func.http.status.404"))
	assert.Equal(t, 1, r.CounterValue("func.http.status.4xx"))
	assert.True(t, r.TimerValue("func.http.time.404") > 0)
	assert.True(t, r.TimerValue("func.http.time.4xx") > 0)
}

func BenchmarkResponseStatsFilter(b *testing.B) {
	r := statstest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost:8080/foo", nil)
	ctx := httpservice.WithServerRequest(context.Background(), NoopResponseWriter{}, req)
	s := saola.Apply(saola.NoopService{}, httpservice.NewResponseStatsFilter(r))
	for i := 0; i < b.N; i++ {
		s.Do(ctx)
	}
}
