package command

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	logger.InitLogger(cfg.Log.Level)

	benchmark, err := app.NewApp(cfg)
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
	case <-shutdown:
		if err = benchmark.Close(); err != nil {
			return err
		}
		slog.Info(fmt.Sprintf("close connection: %v", err))
	}

	slog.Info("app stopped")
	return nil
}
