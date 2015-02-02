package redisservice_test

import (
	"testing"

	"github.com/arjantop/saola"
	"github.com/arjantop/saola/redisservice"
	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/context"
)

type MockConn struct {
	CloseFunc func() error

	ErrFunc func() error

	DoFunc func(commandName string, args ...interface{}) (reply interface{}, err error)

	SendFunc func(commandName string, args ...interface{}) error

	FlushFunc func() error

	ReceiveFunc func() (reply interface{}, err error)
}

func (c *MockConn) Close() error {
	if c.CloseFunc != nil {
		return c.CloseFunc()
	}
	return nil
}

func (c *MockConn) Err() error {
	if c.ErrFunc != nil {
		return c.ErrFunc()
	}
	return nil
}

func (c *MockConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if c.DoFunc != nil {
		return c.DoFunc(commandName, args...)
	}
	return nil, nil
}

func (c *MockConn) Send(commandName string, args ...interface{}) error {
	if c.SendFunc != nil {
		return c.SendFunc(commandName, args...)
	}
	return nil
}

func (c *MockConn) Flush() error {
	if c.FlushFunc != nil {
		return c.FlushFunc()
	}
	return nil
}

func (c *MockConn) Receive() (reply interface{}, err error) {
	if c.ReceiveFunc != nil {
		return c.ReceiveFunc()
	}
	return nil, nil
}

func TestPoolGetClose(t *testing.T) {
	var closeCalled bool
	pool := redisservice.Pool{
		Dial: func() (redis.Conn, error) {
			return &MockConn{
				CloseFunc: func() error {
					closeCalled = true
					return nil
				},
			}, nil
		},
	}
	conn := pool.Get()
	if err := conn.Close(); err != nil {
		t.Error("Close error should be nil: ", err)
	}
	if !closeCalled {
		t.Error("Close on connection should be called")
	}
}

func TestPoolFilter(t *testing.T) {
	var isGetCommand bool
	pool := redisservice.Pool{
		Filter: saola.FuncFilter(func(ctx context.Context, s saola.Service) error {
			req := redisservice.GetClientRequest(ctx)
			req.Command = "GET"
			err := s.Do(ctx)
			req.Response = "response"
			return err
		}),
		Dial: func() (redis.Conn, error) {
			return &MockConn{
				DoFunc: func(cmd string, args ...interface{}) (interface{}, error) {
					isGetCommand = cmd == "GET"
					return nil, nil
				},
			}, nil
		},
	}
	conn := pool.Get()
	r, err := conn.Do(context.Background(), "SET", "key")
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	if rs, ok := r.(string); !ok || rs != "response" {
		t.Error("filter should be executed")
	}
}

func TestPoolClose(t *testing.T) {
	var closedConns int
	pool := redisservice.Pool{
		Dial: func() (redis.Conn, error) {
			return &MockConn{
				CloseFunc: func() error {
					closedConns += 1
					return nil
				},
			}, nil
		},
	}
	pool.Get().Close()
	pool.Get().Close()
	pool.Get()
	if err := pool.Close(); err != nil {
		t.Error("pool should be closed but got: ", err)
	}
	if closedConns != 2 {
		t.Errorf("all the idle connections should be closed but only closed %d", closedConns)
	}
}

func TestPoolGetMaxActive(t *testing.T) {
	pool := redisservice.Pool{
		Dial: func() (redis.Conn, error) {
			return &MockConn{}, nil
		},
		MaxActive: 1,
	}
	pool.Get() // Get one active connection
	conn := pool.Get()
	if conn == nil {
		t.Error("valid connection should always be returned")
	}
	if err := conn.Send(context.Background(), "GET", "key"); err == nil {
		t.Error("connection above MaxActive should always return an error")
	}
}
