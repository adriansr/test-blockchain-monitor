package client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"monitor/pkg/model"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client interface {
	Subscribe(model.Observer[model.Block]) error
	Run(context.Context) error
}

type Opts struct {
	Address        string
	ConnectTimeout time.Duration
}

func New(opts Opts) (Client, error) {
	return &clientImpl{
		address: opts.Address,
		timeout: opts.ConnectTimeout,
	}, nil
}

type clientImpl struct {
	mu        sync.Mutex
	running   bool
	address   string
	timeout   time.Duration
	observers []model.Observer[model.Block]
}

var ErrAlreadyRunning = errors.New("already running")

func (c *clientImpl) Subscribe(obs model.Observer[model.Block]) error {
	if err := c.checkSetRunning(false); err != nil {
		return err
	}
	c.observers = append(c.observers, obs)
	return nil
}

func (c *clientImpl) checkSetRunning(set bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.running {
		return ErrAlreadyRunning
	}
	if set {
		c.running = true
	}
	return nil
}

func (c *clientImpl) Run(ctx context.Context) error {
	if err := c.checkSetRunning(true); err != nil {
		return ErrAlreadyRunning
	}

	cli, err := withTimeout(ctx, c.timeout, ethclient.DialContext, c.address)
	if err != nil {
		return err
	}
	defer cli.Close()

	headC := make(chan *types.Header, 1)
	subsNewHead, err := withTimeout(ctx, c.timeout, cli.SubscribeNewHead, headC)
	if err != nil {
		return err
	}
	defer subsNewHead.Unsubscribe()

	logC := make(chan types.Log, 1)
	subsLogs, err := withTimeout(ctx, c.timeout, subscribeLogs(cli), logC)
	if err != nil {
		return err
	}
	defer subsLogs.Unsubscribe()

	r := NewMerger()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-subsNewHead.Err():
			return fmt.Errorf("reading newHead subscription: %w", err)

		case err := <-subsLogs.Err():
			return fmt.Errorf("reading logs subscription: %w", err)

		case log := <-logC:
			r.AddLog(log)

		case head := <-headC:
			r.AddHead(head)
		}

		for _, blk := range r.Blocks() {
			for _, obs := range c.observers {
				if err := obs.OnEvent(blk); err != nil {
					return fmt.Errorf("observing block #%s: %w", blk.Header.Number.String(), err)
				}
			}
		}
	}
}

func (c *clientImpl) connect(baseCtx context.Context) (*rpc.Client, error) {
	ctx, cancel := context.WithTimeout(baseCtx, c.timeout)
	defer cancel()
	return rpc.DialContext(ctx, c.address)
}

func withTimeout[T, R any](baseCtx context.Context, timeout time.Duration, fn func(context.Context, T) (R, error), param T) (R, error) {
	ctx, cancel := context.WithTimeout(baseCtx, timeout)
	defer cancel()
	return fn(ctx, param)
}

func subscribeLogs(cli *ethclient.Client) func(context.Context, chan<- types.Log) (ethereum.Subscription, error) {
	return func(ctx context.Context, logsC chan<- types.Log) (ethereum.Subscription, error) {
		return cli.SubscribeFilterLogs(ctx, ethereum.FilterQuery{}, logsC)
	}
}
