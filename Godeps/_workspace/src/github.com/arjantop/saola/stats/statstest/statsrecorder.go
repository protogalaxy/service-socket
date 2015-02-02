package statstest

import (
	"time"

	"github.com/arjantop/saola/stats"
)

type StatsRecorder struct {
	scope    string
	counters map[string]int64
	timers   map[string]time.Duration
}

func NewRecorder() *StatsRecorder {
	return &StatsRecorder{
		counters: make(map[string]int64),
		timers:   make(map[string]time.Duration),
	}
}

func (r *StatsRecorder) CounterValue(name string) int64 {
	return r.counters[name]
}

func (r *StatsRecorder) TimerValue(name string) time.Duration {
	return r.timers[name]
}

func (r *StatsRecorder) Counter(name string) stats.Counter {
	return counter{stats.ScopedName(r.scope, name), r.counters}
}

func (r *StatsRecorder) Timer(name string) stats.Timer {
	return timer{stats.ScopedName(r.scope, name), r.timers}
}

func (r *StatsRecorder) Scope(scope string) stats.StatsReceiver {
	return &StatsRecorder{
		scope:    stats.ScopedName(r.scope, scope),
		counters: r.counters,
		timers:   r.timers,
	}
}

type counter struct {
	name     string
	counters map[string]int64
}

func (c counter) Incr() {
	c.Add(1)
}

func (c counter) Add(delta int64) {
	if count, ok := c.counters[c.name]; ok {
		c.counters[c.name] = count + delta
	} else {
		c.counters[c.name] = delta
	}
}

type timer struct {
	name   string
	timers map[string]time.Duration
}

func (t timer) Add(d time.Duration) {
	if time, ok := t.timers[t.name]; ok {
		t.timers[t.name] = time + d
	} else {
		t.timers[t.name] = d
	}
}
