package httpservice

import (
	"fmt"
	"time"

	"github.com/arjantop/saola"
	"golang.org/x/net/context"
)

type LogEntry struct {
	RemoteAddr    string
	RequestTime   time.Time
	RequestMethod string
	RequestPath   string
	StatusCode    int
	Latency       time.Duration
}

func NewRequestLogFilter(f func(e LogEntry)) saola.Filter {
	return saola.FuncFilter(func(ctx context.Context, s saola.Service) error {
		start := time.Now()

		err := s.Do(ctx)

		latency := time.Now().Sub(start)
		req := GetServerRequest(ctx)
		var statusCode int
		if si, ok := req.Writer.(StatusCodeInterceptor); ok {
			statusCode = si.StatusCode()
		}

		entry := LogEntry{
			RemoteAddr:    req.Request.RemoteAddr,
			RequestTime:   start,
			RequestMethod: req.Request.Method,
			RequestPath:   req.Request.URL.Path,
			StatusCode:    statusCode,
			Latency:       latency,
		}
		f(entry)
		return err
	})
}

func NewStdRequestLogFilter() saola.Filter {
	return NewRequestLogFilter(func(e LogEntry) {
		fmt.Printf("%s [%s] \"%s %s\" %d %v\n",
			e.RemoteAddr,
			e.RequestTime,
			e.RequestMethod,
			e.RequestPath,
			e.StatusCode,
			e.Latency)
	})
}
