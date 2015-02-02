package httpservice

import (
	"strconv"
	"time"

	"github.com/arjantop/saola"
	"github.com/arjantop/saola/stats"
	"golang.org/x/net/context"
)

func NewResponseStatsFilter(stats stats.StatsReceiver) saola.Filter {
	return saola.FuncFilter(func(ctx context.Context, s saola.Service) error {
		start := time.Now()
		err := s.Do(ctx)
		latency := time.Now().Sub(start)

		req := GetServerRequest(ctx)

		serviceStats := stats.Scope(s.Name())
		statusStats := serviceStats.Scope("http.status")
		statusTimeStats := serviceStats.Scope("http.time")

		var statusCode int
		if si, ok := req.Writer.(StatusCodeInterceptor); ok {
			statusCode = si.StatusCode()
		}

		statusCodeClass := strconv.Itoa(statusCode/100) + "xx"
		statusCodeStr := strconv.Itoa(statusCode)

		statusStats.Counter(statusCodeStr).Incr()
		statusStats.Counter(statusCodeClass).Incr()

		statusTimeStats.Timer(statusCodeStr).Add(latency)
		statusTimeStats.Timer(statusCodeClass).Add(latency)

		return err
	})
}
