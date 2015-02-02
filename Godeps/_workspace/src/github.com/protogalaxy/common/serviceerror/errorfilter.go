package serviceerror

import (
	"encoding/json"
	"net/http"

	"github.com/arjantop/saola"
	"github.com/arjantop/saola/httpservice"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

func NewErrorResponseFilter() saola.Filter {
	return saola.FuncFilter(func(ctx context.Context, s saola.Service) error {
		err := s.Do(ctx)
		if err != nil {
			req := httpservice.GetServerRequest(ctx)
			req.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")

			if er, ok := err.(ErrorResponse); ok {
				req.Writer.WriteHeader(er.StatusCode)
				encoder := json.NewEncoder(req.Writer)
				encodeError := encoder.Encode(&er)
				if encodeError != nil {
					glog.Warning("error encoding the error response: %s", encodeError)
				}
			} else {
				req.Writer.WriteHeader(http.StatusInternalServerError)
			}
		}
		return err
	})
}

func NewErrorLoggerFilter() saola.Filter {
	return saola.FuncFilter(func(ctx context.Context, s saola.Service) error {
		err := s.Do(ctx)
		if err != nil {
			glog.Errorf("Service error: %s", err)
		}
		return err
	})
}
