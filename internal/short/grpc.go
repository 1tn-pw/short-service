package short

import (
	"context"
	pb "github.com/1tn-pw/protobufs/generated/short_service/v1"
	"github.com/bugfixes/go-bugfixes/utils"
	ConfigBuilder "github.com/keloran/go-config"
)

type Server struct {
	Config ConfigBuilder.Config
	pb.UnimplementedShortServiceServer
}

func (s *Server) CreateURL(ctx context.Context, in *pb.CreateURLRequest) (*pb.CreateURLResponse, error) {
	return &pb.CreateURLResponse{
		Error: utils.StringPointer("not implemented"),
	}, nil
}

func (s *Server) GetURL(ctx context.Context, in *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	return &pb.GetURLResponse{
		Error: utils.StringPointer("not implemented"),
	}, nil
}
