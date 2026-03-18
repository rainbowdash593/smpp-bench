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

type transmitter interface {
	Bind() <-chan smpp.ConnStatus
	Close() error
	Submit(*smpp.ShortMessage) (*smpp.ShortMessage, error)
}

type Connection struct {
	state     <-chan smpp.ConnStatus
	tx        transmitter
	stat      *statistics.Collector
	semaphore chan struct{}
}

func New(cfg *config.Config, limiter *rate.Limiter, collector *statistics.Collector) (*Connection, error) {
	conn := &Connection{
		stat:      collector,
		semaphore: make(chan struct{}, cfg.Connection.WindowSize),
	}
	var tx transmitter
	var txState <-chan smpp.ConnStatus

	if cfg.Connection.BindType == "transceiver" {
		tx = &smpp.Transceiver{
			Addr:        cfg.Connection.Host,
			User:        cfg.Connection.Login,
			Passwd:      cfg.Connection.Password,
			WindowSize:  cfg.Connection.WindowSize,
			RespTimeout: time.Duration(cfg.Connection.ResponseTimeoutMs) * time.Millisecond,
			Handler:     conn.receiveHandler,
			RateLimiter: limiter,
		}
		txState = tx.Bind()
	} else {
		tx = &smpp.Transmitter{
			Addr:        cfg.Connection.Host,
			User:        cfg.Connection.Login,
			Passwd:      cfg.Connection.Password,
			WindowSize:  cfg.Connection.WindowSize,
			RespTimeout: time.Duration(cfg.Connection.ResponseTimeoutMs) * time.Millisecond,
			RateLimiter: limiter,
		}
		rx := &smpp.Receiver{
			Addr:    cfg.Connection.Host,
			User:    cfg.Connection.Login,
			Passwd:  cfg.Connection.Password,
			Handler: conn.receiveHandler,
		}
		txState = tx.Bind()
		_ = rx.Bind()
	}

	var status smpp.ConnStatus

	select {
	case status = <-txState:
		if status.Error() != nil {
			return nil, fmt.Errorf("unable to connect: %v", status.Error())
		}
	case <-time.After(time.Duration(cfg.Connection.BindTimeoutMs) * time.Millisecond):
		return nil, errors.New("timed out waiting for connection")
	}

	conn.tx = tx
	conn.state = txState

	return conn, nil
}

func (c *Connection) receiveHandler(p pdu.Body) {
	c.stat.RecordDLR()
	time.Sleep(5 * time.Millisecond)
	slog.Debug(fmt.Sprintf("read pdu: %s", utils.PDUToString(p)))
}

func (c *Connection) Send(ctx context.Context, message *smpp.ShortMessage) {
	select {
	case c.semaphore <- struct{}{}:
	case <-ctx.Done():
		return
	}

	go func() {
		defer func() { <-c.semaphore }()

		start := time.Now()
		sm, err := c.tx.Submit(message)
		latency := time.Since(start).Milliseconds()
		if err != nil {
			c.stat.RecordFailed(latency)
			slog.Error(fmt.Sprintf("send short message: %v", err))
			return
		}
		c.stat.RecordSuccess(latency)
		slog.Debug("sent message", slog.String("rid", sm.RespID()))
	}()

	return
}

func (c *Connection) C() <-chan smpp.ConnStatus {
	return c.state
}

func (c *Connection) Close() error {
	return c.tx.Close()
}
