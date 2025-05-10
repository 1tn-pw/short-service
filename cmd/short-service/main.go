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
	c := ConfigBuilder.NewConfigNoVault()

	if err := c.Build(ConfigBuilder.Local, ConfigBuilder.Mongo); err != nil {
		_ = logs.Errorf("unable to build config: %v", err)
		return
	}

	if err := service.NewService(*c).Start(); err != nil {
		_ = logs.Errorf("unable to start service: %v", err)
		return
	}
}
