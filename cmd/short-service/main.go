package main

import (
	"github.com/1tn-pw/short-service/internal/service"
	"github.com/bugfixes/go-bugfixes/logs"
	ConfigBuilder "github.com/keloran/go-config"
)

var (
	BuildVersion = "dev"
	BuildHash    = "none"
	ServiceName  = "short-service"
)

func main() {
	logs.Local().Infof("Starting %s %s %s", ServiceName, BuildVersion, BuildHash)

	cfg, err := ConfigBuilder.Build(ConfigBuilder.Local, ConfigBuilder.Vault, ConfigBuilder.Mongo)
	if err != nil {
		_ = logs.Errorf("unable to build config: %v", err)
		return
	}

	if err := service.NewService(*cfg).Start(); err != nil {
		_ = logs.Errorf("unable to start service: %v", err)
		return
	}
}
