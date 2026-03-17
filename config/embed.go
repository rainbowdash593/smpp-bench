package config

import "embed"

//go:embed config.yml
var DefaultConfig embed.FS
