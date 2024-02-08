package service

import (
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	ConfigBuilder "github.com/keloran/go-config"
	"github.com/keloran/go-healthcheck"
	"net/http"
)

type Service struct {
	ConfigBuilder.Config
}

func NewService(cfg ConfigBuilder.Config) *Service {
	return &Service{
		cfg,
	}
}

func (s *Service) Start() error {
	errChan := make(chan error)
	go startHTTP(s.Config, errChan)

	return <-errChan
}

func startHTTP(cfg ConfigBuilder.Config, errChan chan error) {
	logs.Local().Infof("Starting HTTP on %s", cfg.Local.HTTPPort)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthcheck.HTTP)
	errChan <- http.ListenAndServe(fmt.Sprint("%d", cfg.Local.HTTPPort), mux)
}
