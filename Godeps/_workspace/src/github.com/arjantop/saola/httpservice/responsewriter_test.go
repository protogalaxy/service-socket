package httpservice_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arjantop/saola/httpservice"
	"github.com/stretchr/testify/assert"
)

func newTestingResponseWriter() (*httpservice.ResponseWriter, *httptest.ResponseRecorder) {
	r := httptest.NewRecorder()
	return httpservice.NewResponseWriter(r), r
}

func TestResponseWriterDefaultStatusCode(t *testing.T) {
	w, _ := newTestingResponseWriter()
	assert.Equal(t, 200, w.StatusCode())
}

func TestResponseWriterWriteHeader(t *testing.T) {
	w, _ := newTestingResponseWriter()
	w.WriteHeader(http.StatusInternalServerError)
	assert.Equal(t, http.StatusInternalServerError, w.StatusCode())
	w.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusInternalServerError, w.StatusCode())
}

func TestResponseWriterFlush(t *testing.T) {
	w, r := newTestingResponseWriter()
	w.Flush()
	assert.True(t, r.Flushed)
}

type NoopResponseWriter struct{}

func (w NoopResponseWriter) Write(b []byte) (int, error) {
	return 0, nil
}

func (w NoopResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (w NoopResponseWriter) WriteHeader(code int) {}

func TestResponseWriterNonFlushable(t *testing.T) {
	w := httpservice.NewResponseWriter(NoopResponseWriter{})
	assert.NotPanics(t, func() {
		w.Flush()
	})
}

func TestResponseWriterWrite(t *testing.T) {
	w, r := newTestingResponseWriter()
	w.Write([]byte("hello"))
	assert.Equal(t, "hello", r.Body.String())
}

type ClosableResponseWriter struct {
	c chan bool
}

func NewClosableResponseWriter() ClosableResponseWriter {
	return ClosableResponseWriter{
		c: make(chan bool),
	}
}

func (w ClosableResponseWriter) Write(b []byte) (int, error) {
	return 0, nil
}

func (w ClosableResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (w ClosableResponseWriter) WriteHeader(code int) {}

func (w ClosableResponseWriter) CloseNotify() <-chan bool {
	return w.c
}

func (w ClosableResponseWriter) Close() {
	w.c <- true
}

func TestResponseWriterCloseNotify(t *testing.T) {
	cw := NewClosableResponseWriter()
	w := httpservice.NewResponseWriter(cw)
	go func() { cw.Close() }()
	select {
	case c := <-w.CloseNotify():
		assert.True(t, c)
	case <-time.After(time.Microsecond):
		assert.Fail(t, "No close channel response")
	}
}

func TestResponseWriterNonCloseNotifyer(t *testing.T) {
	w := httpservice.NewResponseWriter(NoopResponseWriter{})
	assert.NotPanics(t, func() {
		assert.NotNil(t, w.CloseNotify())
	})
}
