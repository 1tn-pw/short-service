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
	if in.GetUrl() == "" {
		return &pb.CreateURLResponse{
			Error: utils.StringPointer("url is required"),
		}, nil
	}

	resp, err := NewShort(ctx, s.Config).CreateURL(in.Url)
	if err != nil {
		return &pb.CreateURLResponse{
			Error: utils.StringPointer(err.Error()),
		}, nil
	}

	return &pb.CreateURLResponse{
		ShortUrl: resp,
		Url:      in.GetUrl(),
	}, nil
}

func (s *Server) GetURL(ctx context.Context, in *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	if in.GetShortUrl() == "" {
		return &pb.GetURLResponse{
			Error: utils.StringPointer("short url is required"),
		}, nil
	}

	resp, err := NewShort(ctx, s.Config).GetURL(in.ShortUrl)
	if err != nil {
		return &pb.GetURLResponse{
			Error: utils.StringPointer(err.Error()),
		}, nil
	}

	return &pb.GetURLResponse{
		Url: resp,
	}, nil
}
