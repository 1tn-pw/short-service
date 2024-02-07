package service

import ConfigBuilder "github.com/keloran/go-config"

type Service struct {
	ConfigBuilder.Config
}

func NewService(cfg ConfigBuilder.Config) *Service {
	return &Service{
		cfg,
	}
}

func (s *Service) Start() error {
	return nil
}
