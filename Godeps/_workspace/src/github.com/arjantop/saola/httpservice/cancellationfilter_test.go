package httpservice_test

import (
	"testing"
	"time"

	"github.com/arjantop/saola/httpservice"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type SleepService struct{}

func (s SleepService) Do(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Millisecond):
		return nil
	}
}

func (s SleepService) Name() string {
	return "sleep"
}

func TestCancellationFilter_RequestInCancelled(t *testing.T) {
	w := NewClosableResponseWriter()
	ctx := httpservice.WithServerRequest(context.Background(), w, nil)
	go func() { w.Close() }()
	err := httpservice.NewCancellationFilter().Do(ctx, SleepService{})
	assert.Equal(t, context.Canceled, err)
}
