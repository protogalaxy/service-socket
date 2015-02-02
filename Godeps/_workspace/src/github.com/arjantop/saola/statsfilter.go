package saola

import (
	"time"

	"github.com/arjantop/saola/stats"
	"golang.org/x/net/context"
)

func NewStatsFilter(stats stats.StatsReceiver) Filter {
	return FuncFilter(func(ctx context.Context, s Service) error {
		start := time.Now()
		err := s.Do(ctx)
		latency := time.Now().Sub(start)

		serviceStats := stats.Scope(s.Name())
		requestsStat := serviceStats.Counter("requests")
		successStat := serviceStats.Counter("success")
		failureStat := serviceStats.Counter("failure")
		latencyStat := serviceStats.Timer("latency")

		requestsStat.Incr()
		latencyStat.Add(latency)
		if err != nil {
			failureStat.Incr()
		} else {
			successStat.Incr()
		}

		return err
	})
}
