package app

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
	"github.com/pterm/pterm"
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
	output    *pterm.AreaPrinter
	limiter   *rate.Limiter
}

func NewApp(cfg *config.Config, output *pterm.AreaPrinter) (*App, error) {
	limiter := rate.NewLimiter(
		rate.Limit(cfg.Benchmark.InitialRPS),
		max(1, int(float64(cfg.Benchmark.InitialRPS)*cfg.Benchmark.BurstPercentage)),
	)
	collector := statistics.NewCollector()

	conn, err := connection.New(cfg, limiter, collector)
	if err != nil {
		pterm.Error.Printfln("unable to connect to %s", cfg.Connection.Host)
		return nil, err
	}
	pterm.Info.Printfln("connected to %s", cfg.Connection.Host)

	return &App{
		conn:      conn,
		collector: collector,
		cfg:       cfg,
		output:    output,
		limiter:   limiter,
	}, nil
}

func (a *App) TotalStat() statistics.TotalStats {
	return a.collector.Total()
}

func (a *App) Rump() {
	limit := float64(a.limiter.Limit()) * a.cfg.Benchmark.Factor
	current := math.Min(limit, float64(a.cfg.Benchmark.MaxRPS))
	a.limiter.SetLimit(rate.Limit(current))
	a.limiter.SetBurst(max(1, int(current*a.cfg.Benchmark.BurstPercentage)))
	return
}

func (a *App) Run(ctx context.Context) <-chan smpp.ConnStatus {
	pterm.DefaultSection.WithLevel(2).Println("Running benchmark")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				a.conn.Send(ctx, &smpp.ShortMessage{
					Src:  a.cfg.Message.Source,
					Dst:  utils.GeneratePhone(a.cfg.Message.PhoneCountry),
					Text: pdutext.Raw(a.cfg.Message.Text),
				})
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Duration(a.cfg.Benchmark.OutputTickIntervalMs) * time.Millisecond)
		tickCount := 0
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				tickCount += 1
				stat := a.collector.Tick()
				result := fmt.Sprintf(`[Tick №%d] (%d ms)
Rate Limit  = %.1frq/s
Burst       = %d
______________________
RPS         = %.1frq/s
Sent        = %d
Failed      = %d
DLR         = %d
Error Rate  = %.2f%%
Avg Latency = %dms

`,
					tickCount, a.cfg.Benchmark.OutputTickIntervalMs,
					a.limiter.Limit(), a.limiter.Burst(),
					stat.RPS, stat.Sent, stat.Failed, stat.DLR,
					stat.ErrorRate, stat.SentLatencyMs)
				a.output.Update(result)
				slog.Info(fmt.Sprintf("[tick %d] rps=%.1f sent=%d failed=%d dlr=%d error_rate=%.2f%% latency=%dms ",
					tickCount, stat.RPS, stat.Sent, stat.Failed, stat.DLR, stat.ErrorRate, stat.SentLatencyMs))
				a.Rump()
			}
		}
	}()

	return a.conn.C()
}

func (a *App) Close() error {
	return a.conn.Close()
}
