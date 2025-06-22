package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	pb "github.com/1tn-pw/protobufs/generated/short_service/v1"
	"github.com/1tn-pw/short-service/internal/short"
	"github.com/bugfixes/go-bugfixes/logs"
	ConfigBuilder "github.com/keloran/go-config"
	"github.com/keloran/go-healthcheck"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	// Initialize MongoDB connection pool
	if err := short.InitMongo(s.Config); err != nil {
		return logs.Errorf("Failed to initialize MongoDB: %v", err)
	}

	// Ensure MongoDB connection is closed on shutdown
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := short.CloseMongo(ctx); err != nil {
			logs.Errorf("Failed to close MongoDB connection: %v", err)
		}
	}()

	errChan := make(chan error)
	go startHTTP(s.Config, errChan)
	go startGRPC(s.Config, errChan)

	return <-errChan
}

func startHTTP(cfg ConfigBuilder.Config, errChan chan error) {
	logs.Infof("Starting HTTP on %d", cfg.Local.HTTPPort)

	http.HandleFunc("/health", healthcheck.HTTP)
	errChan <- http.ListenAndServe(fmt.Sprintf(":%d", cfg.Local.HTTPPort), nil)
}

func startGRPC(cfg ConfigBuilder.Config, errChan chan error) {
	list, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Local.GRPCPort))
	if err != nil {
		errChan <- err
		return
	}

	gs := grpc.NewServer()
	reflection.Register(gs)
	pb.RegisterShortServiceServer(gs, &short.Server{
		Config: cfg,
	})

	logs.Infof("Starting gRPC on %d", cfg.Local.GRPCPort)
	errChan <- gs.Serve(list)
}
