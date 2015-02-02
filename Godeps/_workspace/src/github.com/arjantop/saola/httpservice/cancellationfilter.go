package httpservice

import (
	"net/http"

	"github.com/arjantop/saola"
	"golang.org/x/net/context"
)

func NewCancellationFilter() saola.Filter {
	return saola.FuncFilter(func(ctx context.Context, s saola.Service) error {
		req := GetServerRequest(ctx)
		if w, ok := req.Writer.(http.CloseNotifier); ok {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			go func() {
				select {
				case <-w.CloseNotify():
					cancel()
				case <-ctx.Done():
					return
				}
			}()
			return s.Do(ctx)
		}
		return s.Do(ctx)
	})
}
