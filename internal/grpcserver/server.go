// grpcserver пакет GRPC сервера
package grpcserver

import (
	"context"
	"database/sql"
	"errors"
	"net"

	"github.com/developerc/GophKeeper/internal/config"
	"github.com/developerc/GophKeeper/internal/entity"
	"github.com/developerc/GophKeeper/internal/entity/myerrors"
	"github.com/developerc/GophKeeper/internal/security"
	"github.com/developerc/GophKeeper/internal/service/dataservice"
	"github.com/developerc/GophKeeper/internal/service/userservice"
	pb "github.com/developerc/GophKeeper/proto"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server структура сервера gRPC
type Server struct {
	pb.GrpcServiceServer
	userService    userservice.UserService
	jwtManager     *security.JWTManager
	storageService dataservice.StorageService
}

// NewServer конструктор сервера gRPC
func NewServer(userService userservice.UserService, jwtManager *security.JWTManager, storageService dataservice.StorageService) *Server {
	return &Server{userService: userService, jwtManager: jwtManager, storageService: storageService}
}

// CreateUser эндпойнт сохранения нового пользователя, генерит токен и отдает в теле респонса
func (s *Server) CreateUser(ctx context.Context, in *pb.UserRegisterRequest) (*pb.AuthorizedResponse, error) {
	login := in.Login
	password := in.Password
	userID := uuid.New().String()

	err := s.userService.Create(ctx, login, password, userID)
	if err != nil {
		var uv *myerrors.UserViolationError
		if errors.As(err, &uv) {
			return nil, status.Errorf(codes.Unauthenticated, "%s", uv.Error())
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	token, err := s.jwtManager.GenerateJWT(userID, login)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb.AuthorizedResponse{Token: token}, nil
}

// LoginUser эндпойнт авторизации существующего пользователя, генерит токен и отдает в теле респонса
func (s *Server) LoginUser(ctx context.Context, in *pb.UserAuthorizedRequest) (*pb.AuthorizedResponse, error) {
	login := in.Login
	password := in.Password

	userID, err := s.userService.Login(ctx, login, password)
	if err != nil {
		var ip *myerrors.InvalidPasswordError
		if errors.As(err, &ip) {
			return nil, status.Errorf(codes.Unauthenticated, "%s", ip.Error())
		}
		return nil, status.Errorf(codes.Internal, "user with login %s not found", login)
	}

	token, err := s.jwtManager.GenerateJWT(userID, login)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb.AuthorizedResponse{Token: token}, nil
}

// SaveRawData эндпойнт сохранения произвольной текстовой информации для авторизованного пользователя
func (s *Server) SaveRawData(ctx context.Context, in *pb.SaveRawDataRequest) (*pb.ErrorResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	//fmt.Println("Comment: ", in.Comment)
	err = s.storageService.SaveRawData(ctx, in.Name, in.Data, userID, in.Comment)
	if err != nil {
		var dv *myerrors.DataViolationError
		if errors.As(err, &dv) {
			return nil, status.Errorf(codes.AlreadyExists, "%s", dv.Error())
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb.ErrorResponse{Error: "no errors"}, nil
}

// SaveLoginWithPassword эндпойнт сохранения логина и пароля для авторизованного пользователя
func (s *Server) SaveLoginWithPassword(ctx context.Context, in *pb.SaveLoginWithPasswordRequest) (*pb.ErrorResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	//fmt.Println("Comment: ", in.Comment)
	err = s.storageService.SaveLoginWithPassword(ctx, in.Name, in.Login, in.Password, userID, in.Comment)
	if err != nil {
		var dv *myerrors.DataViolationError
		if errors.As(err, &dv) {
			return nil, status.Errorf(codes.AlreadyExists, "%s", dv.Error())
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb.ErrorResponse{Error: "no errors"}, nil
}

// SaveBinaryData эндпойнт сохранения произвольных бинарных данных для авторизованного пользователя
func (s *Server) SaveBinaryData(ctx context.Context, in *pb.SaveBinaryDataRequest) (*pb.ErrorResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	//fmt.Println("Comment: ", in.Comment)
	err = s.storageService.SaveBinaryData(ctx, in.Name, in.Data, userID, in.Comment)
	if err != nil {
		var dv *myerrors.DataViolationError
		if errors.As(err, &dv) {
			return nil, status.Errorf(codes.AlreadyExists, "%s", dv.Error())
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb.ErrorResponse{Error: "no errors"}, nil
}

// SaveCardData эндпойнт сохранения данных банковской карты для авторизованного пользователя
func (s *Server) SaveCardData(ctx context.Context, in *pb.SaveCardDataRequest) (*pb.ErrorResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	//fmt.Println("Comment: ", in.Comment)

	card := entity.CardDataDTO{
		Number:     in.Number,
		Month:      in.Month,
		Year:       in.Year,
		CardHolder: in.CardHolder,
		Cvv:        in.Cvv,
	}

	err = s.storageService.SaveCardData(ctx, in.Name, card, userID, in.Comment)
	if err != nil {
		var dv *myerrors.DataViolationError
		if errors.As(err, &dv) {
			return nil, status.Errorf(codes.AlreadyExists, "%s", dv.Error())
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb.ErrorResponse{Error: "no errors"}, nil
}

// GetRawData эндпойнт получения текстовой информации по названию для авторизованного пользователя
func (s *Server) GetRawData(ctx context.Context, in *pb.GetRawDataRequest) (*pb.GetRawDataResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}

	data, comment, err := s.storageService.GetRawData(ctx, in.Name, userID)
	if err != nil {
		var nf *myerrors.NotFoundError
		if errors.As(err, &nf) {
			return nil, status.Errorf(codes.NotFound, "%s", nf.Error())
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	//fmt.Println(comment)

	return &pb.GetRawDataResponse{Data: data, Comment: comment}, nil
}

// GetLoginWithPassword эндпойнт получения логина и пароля по названию для авторизованного пользователя
func (s *Server) GetLoginWithPassword(ctx context.Context, in *pb.GetLoginWithPasswordRequest) (*pb.GetLoginWithPasswordResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}

	data, comment, err := s.storageService.GetLoginWithPassword(ctx, in.Name, userID)
	if err != nil {
		var nf *myerrors.NotFoundError
		if errors.As(err, &nf) {
			return nil, status.Errorf(codes.NotFound, "%s", nf.Error())
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	//fmt.Println(comment)

	return &pb.GetLoginWithPasswordResponse{Login: data.Login, Password: data.Password, Comment: comment}, nil
}

// GetBinaryData эндпойнт получения произвольных бинарных данных по названию для авторизованного пользователя
func (s *Server) GetBinaryData(ctx context.Context, in *pb.GetBinaryDataRequest) (*pb.GetBinaryDataResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}

	data, comment, err := s.storageService.GetBinaryData(ctx, in.Name, userID)
	if err != nil {
		var nf *myerrors.NotFoundError
		if errors.As(err, &nf) {
			return nil, status.Errorf(codes.NotFound, "%s", nf.Error())
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	//fmt.Println(comment)

	return &pb.GetBinaryDataResponse{Data: data, Comment: comment}, nil
}

// GetCardData эндпойнт получения данных банковской карты по названию для авторизованного пользователя
func (s *Server) GetCardData(ctx context.Context, in *pb.GetCardDataRequest) (*pb.GetCardDataResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}

	data, comment, err := s.storageService.GetCardData(ctx, in.Name, userID)
	if err != nil {
		var nf *myerrors.NotFoundError
		if errors.As(err, &nf) {
			return nil, status.Errorf(codes.NotFound, "%s", nf.Error())
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	//fmt.Println(comment)

	return &pb.GetCardDataResponse{Number: data.Number, Month: data.Month, Year: data.Year, CardHolder: data.CardHolder, Cvv: data.Cvv, Comment: comment}, nil
}

// GetAllSavedDataNames метод для получения всех названий сохранений
func (s *Server) GetAllSavedDataNames(ctx context.Context, in *pb.GetAllSavedDataNamesRequest) (*pb.GetAllSavedDataNamesResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}

	names, err := s.storageService.GetAllSavedDataNames(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb.GetAllSavedDataNamesResponse{SavedDataNames: names}, nil
}

// DelRawData удаляет сырые данные
func (s *Server) DelRawData(ctx context.Context, in *pb.DelRequest) (*pb.ErrorResponse, error) {
	config.ServerSettingsGlob.Logger.Info("DelRawData", zap.String("server", "delete data from db"))
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	err = s.storageService.DelDataByNameUserId(ctx, in.Name, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &pb.ErrorResponse{Error: "no errors"}, nil
}

// DelLoginWithPassword удаляет данные логин, пароль
func (s *Server) DelLoginWithPassword(ctx context.Context, in *pb.DelRequest) (*pb.ErrorResponse, error) {
	config.ServerSettingsGlob.Logger.Info("DelLoginWithPassword", zap.String("server", "delete data from db"))
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	err = s.storageService.DelDataByNameUserId(ctx, in.Name, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &pb.ErrorResponse{Error: "no errors"}, nil
}

// DelBinaryData удаляет бинарные данные
func (s *Server) DelBinaryData(ctx context.Context, in *pb.DelRequest) (*pb.ErrorResponse, error) {
	config.ServerSettingsGlob.Logger.Info("DelBinaryData", zap.String("server", "delete data from db"))
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	err = s.storageService.DelDataByNameUserId(ctx, in.Name, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &pb.ErrorResponse{Error: "no errors"}, nil
}

// DelCardData удаляет данные карты
func (s *Server) DelCardData(ctx context.Context, in *pb.DelRequest) (*pb.ErrorResponse, error) {
	config.ServerSettingsGlob.Logger.Info("DelCardData", zap.String("server", "delete data from db"))
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	err = s.storageService.DelDataByNameUserId(ctx, in.Name, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &pb.ErrorResponse{Error: "no errors"}, nil
}

// UpdRawData обновляет сырые данные
func (s *Server) UpdRawData(ctx context.Context, in *pb.SaveRawDataRequest) (*pb.ErrorResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	//fmt.Println("Comment: ", in.Comment)
	err = s.storageService.UpdRawData(ctx, in.Name, in.Data, userID, in.Comment)
	if err != nil {
		return nil, err
	}

	return &pb.ErrorResponse{Error: "no errors"}, nil
}

// UpdLoginWithPassword обновляет данные логин, пароль
func (s *Server) UpdLoginWithPassword(ctx context.Context, in *pb.SaveLoginWithPasswordRequest) (*pb.ErrorResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	//fmt.Println("Comment: ", in.Comment)
	err = s.storageService.UpdLoginWithPassword(ctx, in.Name, in.Login, in.Password, userID, in.Comment)
	if err != nil {
		return nil, err
	}

	return &pb.ErrorResponse{Error: "no errors"}, nil
}

// UpdBinaryData обновляет бинарные данные
func (s *Server) UpdBinaryData(ctx context.Context, in *pb.SaveBinaryDataRequest) (*pb.ErrorResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	//fmt.Println("Comment: ", in.Comment)
	err = s.storageService.UpdBinaryData(ctx, in.Name, in.Data, userID, in.Comment)
	if err != nil {
		return nil, err
	}

	return &pb.ErrorResponse{Error: "no errors"}, nil
}

// UpdCardData обновляет данные карты
func (s *Server) UpdCardData(ctx context.Context, in *pb.SaveCardDataRequest) (*pb.ErrorResponse, error) {
	userID, err := s.jwtManager.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	//fmt.Println("Comment: ", in.Comment)
	card := entity.CardDataDTO{
		Number:     in.Number,
		Month:      in.Month,
		Year:       in.Year,
		CardHolder: in.CardHolder,
		Cvv:        in.Cvv,
	}

	err = s.storageService.UpdCardData(ctx, in.Name, card, userID, in.Comment)
	if err != nil {
		return nil, err
	}

	return &pb.ErrorResponse{Error: "no errors"}, nil
}

// NewGRPCserver конструктор GRPC сервера
func NewGRPCserver(ctx context.Context, settings *config.ServerSettings, userService userservice.UserService, jwtManager *security.JWTManager, storageService dataservice.StorageService, db *sql.DB) {
	lis, err := net.Listen("tcp", settings.Host) // будем ждать запросы по этому адресу
	if err != nil {
		settings.Logger.Info("Init gRPC service", zap.String("error", err.Error()))
		return
	}
	settings.Logger.Info("Init gRPC service", zap.String("start at host:port", settings.Host))

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(security.ServerJwtInterceptor))

	go func() {
		defer close(beforeStop)
		<-ctx.Done()
		settings.Logger.Info("Server gRPC", zap.String("shutdown", "begin"))
		grpcServer.GracefulStop()
		settings.Logger.Info("Server gRPC", zap.String("shutdown", "end"))
		db.Close()
		settings.Logger.Info("DB", zap.String("close", "end"))
	}()

	reductorServiceServer := NewServer(userService, jwtManager, storageService)

	pb.RegisterGrpcServiceServer(grpcServer, reductorServiceServer)
	if err := grpcServer.Serve(lis); err != nil {
		settings.Logger.Info("Init gRPC service", zap.String("error", err.Error()))
		return
	}
}
