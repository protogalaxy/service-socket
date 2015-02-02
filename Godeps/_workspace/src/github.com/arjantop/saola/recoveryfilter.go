package saola

import (
	"fmt"

	"golang.org/x/net/context"
)

func NewRecoveryFilter() Filter {
	return FuncFilter(func(ctx context.Context, s Service) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %s", r)
			}
		}()
		return s.Do(ctx)
	})
}
