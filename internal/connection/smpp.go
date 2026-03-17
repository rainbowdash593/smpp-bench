package connection

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu"
	"github.com/rainbowdash593/smpp-bench/config"
	"github.com/rainbowdash593/smpp-bench/internal/statistics"
	"github.com/rainbowdash593/smpp-bench/pkg/utils"
	"golang.org/x/time/rate"
)

type Connection struct {
	state     <-chan smpp.ConnStatus
	trx       *smpp.Transceiver
	stat      *statistics.Collector
	semaphore chan struct{}
}

func New(cfg *config.Config, limiter *rate.Limiter, collector *statistics.Collector) (*Connection, error) {
	conn := &Connection{
		stat:      collector,
		semaphore: make(chan struct{}, cfg.Connection.WindowSize),
	}
	trx := &smpp.Transceiver{
		Addr:        cfg.Connection.Host,
		User:        cfg.Connection.Login,
		Passwd:      cfg.Connection.Password,
		WindowSize:  cfg.Connection.WindowSize,
		Handler:     conn.receiveHandler,
		RateLimiter: limiter,
	}
	state := trx.Bind()
	var status smpp.ConnStatus

	select {
	case status = <-state:
		if status.Error() != nil {
			return nil, fmt.Errorf("unable to connect: %v", status.Error())
		}
	case <-time.After(time.Duration(cfg.Connection.BindTimeoutMs) * time.Millisecond):
		return nil, errors.New("timed out waiting for connection")
	}

	conn.trx = trx
	conn.state = state

	return conn, nil
}

func (c *Connection) receiveHandler(p pdu.Body) {
	c.stat.RecordDLR()
	slog.Debug(fmt.Sprintf("read pdu: %s", utils.PDUToString(p)))
}

func (c *Connection) Send(ctx context.Context, message *smpp.ShortMessage) error {
	done := make(chan error)

	select {
	case c.semaphore <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	go func() {
		defer func() { <-c.semaphore }()

		start := time.Now()
		sm, err := c.trx.Submit(message)
		latency := time.Since(start).Milliseconds()
		if err != nil {
			c.stat.RecordFailed(latency)
			done <- err
			return
		}
		c.stat.RecordSuccess(latency)
		slog.Debug("sent message", slog.String("rid", sm.RespID()))
		done <- nil
	}()

	return <-done
}

func (c *Connection) C() <-chan smpp.ConnStatus {
	return c.state
}

func (c *Connection) Close() error {
	return c.trx.Close()
}
