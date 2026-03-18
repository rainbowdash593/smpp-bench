package command

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pterm/pterm"
	"github.com/rainbowdash593/smpp-bench/config"
	"github.com/rainbowdash593/smpp-bench/internal/app"
	"github.com/rainbowdash593/smpp-bench/pkg/logger"
)

type RunCmd struct {
	Path string `arg:"" optional:"" name:"path" help:"configuration file path" type:"path"`
}

func (c *RunCmd) readConfig() (*config.Config, error) {
	configPath := "./bench_config.yml"
	if c.Path != "" {
		configPath = c.Path
	}
	return config.NewConfig(configPath)
}

func (c *RunCmd) Run() error {
	cfg, err := c.readConfig()
	if err != nil {
		return err
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	err = logger.InitLogger(cfg.Log.Level, cfg.Log.Enabled, cfg.Log.Filename)
	if err != nil {
		return err
	}
	pterm.Info.Println("starting benchmark")
	area, err := pterm.DefaultArea.Start()
	if err != nil {
		return err
	}
	benchmark, err := app.NewApp(cfg, area)
	if err != nil {
		return err
	}
	slog.Info("app started")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stateCh := benchmark.Run(ctx)
	select {
	case state := <-stateCh:
		slog.Info(fmt.Sprintf("connection status: %s, err: %s", state.Status(), state.Error()))
		if state.Error() != nil {
			pterm.Error.Printfln("Connection error: %s", state.Error())
		} else {
			pterm.Info.Printfln("connection status: %s, err: %s", state.Status(), state.Error())
		}
	case <-shutdown:
		if err = benchmark.Close(); err != nil {
			return err
		}
		pterm.Info.Println("Benchmark stopped")
		slog.Info(fmt.Sprintf("close connection: %v", err))
	}
	total := benchmark.TotalStat()
	tableData := pterm.TableData{
		{"Sent", "Failed", "DLR", "Error Rate", "Avg Latency", "Window Size", "Avg RPS", "Max RPS"},
		{
			fmt.Sprintf("%d", total.Sent),
			fmt.Sprintf("%d", total.Failed),
			fmt.Sprintf("%d", total.DLR),
			fmt.Sprintf("%.2f%%", total.ErrorRate),
			fmt.Sprintf("%dms", total.SentLatencyMs),
			fmt.Sprintf("%d", cfg.Connection.WindowSize),
			fmt.Sprintf("%.1frq/s", total.AvgRPS),
			fmt.Sprintf("%.1frq/s", total.MaxRPS),
		},
	}
	pterm.DefaultSection.WithLevel(2).Println("Total statistics")
	_ = pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()

	slog.Info("app stopped")
	return nil
}
