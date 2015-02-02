package httpservice

import "net/http"

type StatusCodeInterceptor interface {
	StatusCode() int
}

type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, 0}
}

func (w *ResponseWriter) WriteHeader(code int) {
	if w.statusCode == 0 {
		w.statusCode = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *ResponseWriter) StatusCode() int {
	if w.statusCode == 0 {
		return 200
	}
	return w.statusCode
}

func (w *ResponseWriter) CloseNotify() <-chan bool {
	if n, ok := w.ResponseWriter.(http.CloseNotifier); ok {
		return n.CloseNotify()
	}
	c := make(chan bool)
	return c
}

func (w *ResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
