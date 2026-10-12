package natsx

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/Jordan-J-Brown/casino-go/pkg/chans"
	"github.com/nats-io/nats.go"
)

type ConnPool struct {
	conns chan *nats.Conn
}

type ConnPoolConfig struct {
	URL      string //Defaults to nats.DefaultURL
	MaxConns int    //Defaults to runtime.NumCPU
	Opts     []nats.Option
}

func NewConnPool(ctx context.Context, cfg ConnPoolConfig) (*ConnPool, error) {
	if cfg.URL == "" {
		cfg.URL = nats.DefaultURL
	}
	if cfg.MaxConns == 0 {
		cfg.MaxConns = runtime.NumCPU()
	}

	pool := &ConnPool{
		conns: make(chan *nats.Conn, cfg.MaxConns),
	}

	for range cfg.MaxConns {
		conn, err := nats.Connect(cfg.URL, cfg.Opts...)
		if err != nil {
			return nil, err
		}
		pool.conns <- conn
	}

	return pool, nil
}

func (pool *ConnPool) RequestMsg(ctx context.Context, msg *nats.Msg, timeout time.Duration) (*nats.Msg, error) {
	var conn *nats.Conn
	select {
	case <-ctx.Done():
		return nil, errors.New("timeout exceeded whilst waiting for nats connection from pool")
	case conn = <-pool.conns:
	}
	defer func() {
		pool.conns <- conn
	}()

	reply, err := conn.RequestMsg(msg, timeout)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func (pool *ConnPool) Close() error {
	close(pool.conns)
	conns := chans.Collect(pool.conns)

	failedDrains := make([]error, 0, len(pool.conns))
	for _, conn := range conns {
		if conn != nil {
			if err := conn.Drain(); err != nil {
				failedDrains = append(failedDrains, err)
			}
		}
	}

	if len(failedDrains) > 0 {
		return errors.Join(append([]error{errors.New("failed to drain pooled connections:")}, failedDrains...)...)
	}

	return nil
}
