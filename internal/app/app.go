package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
	"github.com/rainbowdash593/smpp-bench/config"
	"github.com/rainbowdash593/smpp-bench/internal/connection"
	"github.com/rainbowdash593/smpp-bench/internal/statistics"
	"github.com/rainbowdash593/smpp-bench/pkg/utils"
	"golang.org/x/time/rate"
)

type App struct {
	conn      *connection.Connection
	collector *statistics.Collector
	cfg       *config.Config
}

func NewApp(cfg *config.Config) (*App, error) {
	limiter := rate.NewLimiter(10, 1)
	collector := statistics.NewCollector()
	conn, err := connection.New(cfg, limiter, collector)
	if err != nil {
		return nil, err
	}
	return &App{
		conn:      conn,
		collector: collector,
		cfg:       cfg,
	}, nil
}

func (a *App) Run(ctx context.Context) <-chan smpp.ConnStatus {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := a.conn.Send(ctx, &smpp.ShortMessage{
					Src:  "VIRTA",
					Dst:  utils.GeneratePhone("RU"),
					Text: pdutext.Raw("Hello benchmark!"),
				}); err != nil {
					slog.Error(fmt.Sprintf("send short message: %v", err))
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stat := a.collector.Tick()
				slog.Info(fmt.Sprintf("[tick] rps=%.1f sent=%d failed=%d dlr=%d error_rate=%.2f%% latency=%dms ",
					stat.RPS, stat.Sent, stat.Failed, stat.DLR, stat.ErrorRate, stat.SentLatencyMs))
			}
		}
	}()

	return a.conn.C()
}

func (a *App) Close() error {
	return a.conn.Close()
}
