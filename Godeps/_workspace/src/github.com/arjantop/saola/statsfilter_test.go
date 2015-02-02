package saola_test

import (
	"errors"
	"testing"

	"github.com/arjantop/saola"
	"github.com/arjantop/saola/stats/statstest"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestStatsFilter(t *testing.T) {
	r := statstest.NewRecorder()
	f := saola.NewStatsFilter(r)
	err := f.Do(context.Background(), saola.FuncService(func(ctx context.Context) error {
		return nil
	}))
	assert.NoError(t, err)
	assert.Equal(t, 1, r.CounterValue("func.requests"))
	assert.Equal(t, 1, r.CounterValue("func.success"))
	assert.Equal(t, 0, r.CounterValue("func.failure"))
	assert.True(t, r.TimerValue("func.latency") > 0)

	err = f.Do(context.Background(), saola.FuncService(func(ctx context.Context) error {
		return errors.New("error")
	}))
	assert.Error(t, err)
	assert.Equal(t, 2, r.CounterValue("func.requests"))
	assert.Equal(t, 1, r.CounterValue("func.success"))
	assert.Equal(t, 1, r.CounterValue("func.failure"))
	assert.True(t, r.TimerValue("func.latency") > 0)
}

func BenchmarkStatsFilter(b *testing.B) {
	r := statstest.NewRecorder()
	ctx := context.Background()
	s := saola.Apply(saola.NoopService{}, saola.NewStatsFilter(r))
	for i := 0; i < b.N; i++ {
		s.Do(ctx)
	}
}
