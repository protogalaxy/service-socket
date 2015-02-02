package saola

import "golang.org/x/net/context"

type Filter interface {
	Do(ctx context.Context, s Service) error
}

type FuncFilter func(ctx context.Context, s Service) error

func (f FuncFilter) Do(ctx context.Context, s Service) error {
	return f(ctx, s)
}

func Chain(f Filter, fs ...Filter) Filter {
	if len(fs) == 0 {
		return f
	} else {
		chained := Chain(fs[0], fs[1:]...)
		return FuncFilter(func(ctx context.Context, s Service) error {
			return f.Do(ctx, FuncService(func(ctx context.Context) error {
				return chained.Do(ctx, s)
			}))
		})
	}
}

type Service interface {
	Do(ctx context.Context) error
	Name() string
}

type FuncService func(ctx context.Context) error

func (f FuncService) Do(ctx context.Context) error {
	return f(ctx)
}

func (f FuncService) Name() string {
	return "func"
}

type filteredService struct {
	original Service
	filter   Filter
}

func (s filteredService) Do(ctx context.Context) error {
	return s.filter.Do(ctx, s.original)
}

func (s filteredService) Name() string {
	return s.original.Name()
}

func Apply(s Service, fs ...Filter) Service {
	if len(fs) == 0 {
		return s
	} else {
		f := fs[0]
		s := Apply(s, fs[1:]...)
		return filteredService{s, f}
	}
}

type NoopService struct{}

func (s NoopService) Do(ctx context.Context) error {
	return nil
}

func (s NoopService) Name() string {
	return "noop"
}
