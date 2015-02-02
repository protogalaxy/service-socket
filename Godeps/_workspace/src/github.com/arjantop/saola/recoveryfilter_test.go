package saola_test

import (
	"errors"
	"testing"

	"github.com/arjantop/saola"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func recoveryService(action string) saola.Service {
	return saola.Apply(saola.FuncService(func(ctx context.Context) error {
		switch action {
		case "error":
			return errors.New("error")
		case "panic":
			panic("original")
		default:
			return nil
		}
	}), saola.NewRecoveryFilter())
}

func TestRecoverFilterPanic(t *testing.T) {
	s := recoveryService("panic")
	err := s.Do(context.Background())
	assert.Equal(t, errors.New("panic: original"), err)
}

func TestRecoverFilterError(t *testing.T) {
	s := recoveryService("error")
	err := s.Do(context.Background())
	assert.Equal(t, errors.New("error"), err, "Only in the case of a panic there are be side-effects")
}
