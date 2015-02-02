package statstest_test

import (
	"testing"
	"time"

	"github.com/arjantop/saola/stats/statstest"
	"github.com/stretchr/testify/assert"
)

func TestStatsRecorderCounter(t *testing.T) {
	r := statstest.NewRecorder()
	c1 := r.Counter("a.b")
	c1.Incr()
	assert.Equal(t, 1, r.CounterValue("a.b"))
	c2 := r.Counter("a.b.c")
	c2.Add(2)
	assert.Equal(t, 2, r.CounterValue("a.b.c"))
	c2.Add(-1)
	assert.Equal(t, 1, r.CounterValue("a.b.c"))
	c1.Incr()
	assert.Equal(t, 2, r.CounterValue("a.b"))
}

func TestStatsRecorderMultipleCounterInstances(t *testing.T) {
	r := statstest.NewRecorder()
	c1 := r.Counter("a")
	c2 := r.Counter("a")
	c1.Incr()
	assert.Equal(t, 1, r.CounterValue("a"))
	c2.Incr()
	assert.Equal(t, 2, r.CounterValue("a"))
}

func TestStatsRecorderTimer(t *testing.T) {
	r := statstest.NewRecorder()
	c1 := r.Timer("a.b")
	c1.Add(time.Millisecond)
	assert.Equal(t, time.Millisecond, r.TimerValue("a.b"))
	c2 := r.Timer("a.b.c")
	c2.Add(5 * time.Second)
	assert.Equal(t, 5*time.Second, r.TimerValue("a.b.c"))
	c2.Add(time.Millisecond)
	assert.Equal(t, 5001*time.Millisecond, r.TimerValue("a.b.c"))
	c1.Add(-time.Second)
	assert.Equal(t, -999*time.Millisecond, r.TimerValue("a.b"))
}

func TestStatsRecorderMultipleTimerInstances(t *testing.T) {
	r := statstest.NewRecorder()
	c1 := r.Timer("a")
	c2 := r.Timer("a")
	c1.Add(time.Second)
	assert.Equal(t, time.Second, r.TimerValue("a"))
	c2.Add(time.Minute)
	assert.Equal(t, 61*time.Second, r.TimerValue("a"))
}

func TestStatsRecorderScope(t *testing.T) {
	r := statstest.NewRecorder()
	s1 := r.Scope("a")
	c1 := s1.Counter("b")
	c1.Incr()
	assert.Equal(t, 1, r.CounterValue("a.b"))

	t1 := s1.Timer("b")
	t1.Add(time.Second)
	assert.Equal(t, time.Second, r.TimerValue("a.b"))
}

func TestStatsRecorderNestedScope(t *testing.T) {
	r := statstest.NewRecorder()
	s1 := r.Scope("a").Scope("b")
	c1 := s1.Counter("c")
	c1.Incr()
	assert.Equal(t, 1, r.CounterValue("a.b.c"))

	t1 := s1.Timer("c")
	t1.Add(time.Second)
	assert.Equal(t, time.Second, r.TimerValue("a.b.c"))
}
