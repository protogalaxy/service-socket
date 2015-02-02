package redisservice

import (
	"io"
	"sync"
	"time"

	"github.com/arjantop/saola"
	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/context"
)

type Client interface {
	Do(ctx context.Context, cmd string, args ...interface{}) (interface{}, error)
	Send(ctx context.Context, cmd string, args ...interface{}) error
	io.Closer
}

type Pool struct {
	Filter  saola.Filter
	service saola.Service

	Dial func() (redis.Conn, error)

	TestOnBorrow func(redis.Conn, time.Time) error

	MaxIdle   int
	MaxActive int

	IdleTimeout time.Duration

	lock     sync.Mutex
	implPool *redis.Pool
}

func (p *Pool) pool() *redis.Pool {
	p.lock.Lock()
	if p.implPool == nil {
		p.service = redisService{}
		if p.Filter != nil {
			p.service = saola.Apply(p.service, p.Filter)
		}

		p.implPool = &redis.Pool{
			Dial:         p.Dial,
			TestOnBorrow: p.TestOnBorrow,
			MaxIdle:      p.MaxIdle,
			MaxActive:    p.MaxActive,
			IdleTimeout:  p.IdleTimeout,
		}
	}
	result := p.implPool
	p.lock.Unlock()
	return result
}

func (p *Pool) Close() error {
	return p.implPool.Close()
}

func (p *Pool) Get() Client {
	return &connClient{
		service: p.service,
		conn:    p.pool().Get(),
	}
}

type ClientRequest struct {
	conn        redis.Conn
	requestType requestType
	Command     string
	Args        []interface{}
	Response    interface{}
}

func GetClientRequest(ctx context.Context) *ClientRequest {
	req, _ := ctx.Value(requestKey{}).(*ClientRequest)
	return req
}

type requestKey struct{}

type redisService struct{}

func (rs redisService) Do(ctx context.Context) error {
	req := GetClientRequest(ctx)
	switch req.requestType {
	case Do:
		r, err := req.conn.Do(req.Command, req.Args...)
		req.Response = r
		return err
	case Send:
		err := req.conn.Send(req.Command, req.Args...)
		return err
	}
	return nil
}

func (rs redisService) Name() string {
	return "redis"
}

type connClient struct {
	service saola.Service

	conn redis.Conn
}

type requestType int

const (
	Do   requestType = iota
	Send requestType = iota
)

func (c *connClient) Do(ctx context.Context, cmd string, args ...interface{}) (interface{}, error) {
	r := &ClientRequest{
		conn:        c.conn,
		requestType: Do,
		Command:     cmd,
		Args:        args,
	}
	err := c.service.Do(context.WithValue(ctx, requestKey{}, r))
	return r.Response, err
}

func (c *connClient) Send(ctx context.Context, cmd string, args ...interface{}) error {
	r := &ClientRequest{
		conn:        c.conn,
		requestType: Send,
		Command:     cmd,
		Args:        args,
	}
	err := c.service.Do(context.WithValue(ctx, requestKey{}, r))
	return err
}

func (c *connClient) Close() error {
	return c.conn.Close()
}
