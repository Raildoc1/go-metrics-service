package server

import (
	"fmt"
	"go-metrics-service/internal/server/grpcservers"
	pb "go-metrics-service/proto"
	"net"

	"google.golang.org/grpc"
)

var _ pb.UpdateMetricsServer = (*GRPCServer)(nil)

type GRPCServer struct {
	pb.UnimplementedUpdateMetricsServer
	cfg        GRPCConfig
	controller GRPCController
	server     *grpc.Server
}

type GRPCController interface {
	grpcservers.Controller
}

type GRPCConfig struct {
	Port uint16
}

func NewGRPC(cfg GRPCConfig, controller GRPCController) *GRPCServer {
	return &GRPCServer{
		controller: controller,
		server:     grpc.NewServer(),
		cfg:        cfg,
	}
}

func (s *GRPCServer) Run() error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", s.cfg.Port))
	if err != nil {
		return fmt.Errorf("failed to start listen: %w", err)
	}

	ums := grpcservers.NewUpdateMetricsServer(s.controller)

	pb.RegisterUpdateMetricsServer(s.server, ums)

	if err := s.server.Serve(listen); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *GRPCServer) Shutdown() {
	s.server.GracefulStop()
}
