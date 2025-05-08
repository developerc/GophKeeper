package grpcserver

import (
	"context"
	"fmt"
	"net"

	"github.com/developerc/GophKeeper/internal/config"
	pb "github.com/developerc/GophKeeper/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Server структура сервера gRPC
type Server struct {
	pb.GrpcServiceServer
	//Service svc
}

// NewServer конструктор сервера gRPC
func NewServer() *Server {
	return &Server{}
}

func (s *Server) CreateUser(ctx context.Context, in *pb.UserRegisterRequest) (*pb.AuthorizedResponse, error) {
	fmt.Println(in.Login, in.Password)
	return &pb.AuthorizedResponse{Token: "token"}, nil
}

func (s *Server) LoginUser(ctx context.Context, in *pb.UserAuthorizedRequest) (*pb.AuthorizedResponse, error) {
	return nil, nil
}

func (s *Server) SaveRawData(ctx context.Context, in *pb.SaveRawDataRequest) (*pb.ErrorResponse, error) {
	return nil, nil
}

func (s *Server) SaveLoginWithPassword(ctx context.Context, in *pb.SaveLoginWithPasswordRequest) (*pb.ErrorResponse, error) {
	return nil, nil
}

func (s *Server) SaveBinaryData(ctx context.Context, in *pb.SaveBinaryDataRequest) (*pb.ErrorResponse, error) {
	return nil, nil
}

func (s *Server) SaveCardData(ctx context.Context, in *pb.SaveCardDataRequest) (*pb.ErrorResponse, error) {
	return nil, nil
}

func (s *Server) GetRawData(ctx context.Context, in *pb.GetRawDataRequest) (*pb.GetRawDataResponse, error) {
	return nil, nil
}

func (s *Server) GetLoginWithPassword(ctx context.Context, in *pb.GetLoginWithPasswordRequest) (*pb.GetLoginWithPasswordResponse, error) {
	return nil, nil
}

func (s *Server) GetBinaryData(ctx context.Context, in *pb.GetBinaryDataRequest) (*pb.GetBinaryDataResponse, error) {
	return nil, nil
}

func (s *Server) GetCardData(ctx context.Context, in *pb.GetCardDataRequest) (*pb.GetCardDataResponse, error) {
	return nil, nil
}

func (s *Server) GetAllSavedDataNames(ctx context.Context, in *pb.GetAllSavedDataNamesRequest) (*pb.GetAllSavedDataNamesResponse, error) {
	return nil, nil
}

func NewGRPCserver(ctx context.Context, settings *config.ServerSettings) {
	lis, err := net.Listen("tcp", settings.Host) // будем ждать запросы по этому адресу
	if err != nil {
		settings.Logger.Info("Init gRPC service", zap.String("error", err.Error()))
		return
	}
	settings.Logger.Info("Init gRPC service", zap.String("start at host:port", settings.Host))

	grpcServer := grpc.NewServer()
	reductorServiceServer := NewServer()

	pb.RegisterGrpcServiceServer(grpcServer, reductorServiceServer)
	if err := grpcServer.Serve(lis); err != nil {
		settings.Logger.Info("Init gRPC service", zap.String("error", err.Error()))
		return
	}

}
