package command

import (
	"io/fs"
	"os"

	"github.com/rainbowdash593/smpp-bench/config"
)

type InitCmd struct {
	Path string `arg:"" optional:"" name:"path" help:"the path where the configuration file will be generated" type:"path"`
}

func (c *InitCmd) Run() error {
	data, err := fs.ReadFile(config.DefaultConfig, "config.yml")
	if err != nil {
		return err
	}
	configPath := "./bench_config.yml"
	if c.Path != "" {
		configPath = c.Path
	}
	if err = os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}
	return nil
}
