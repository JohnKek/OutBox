package api

import (
	"context"
	"fmt"
	api "github.com/JohnKek/OutBox/api/api/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"sync"
)

type Server struct {
	api.U
	sync.Mutex
}

func (s *Server) GetPerson(_ context.Context, request *api.GetPersonRequest) (*api.PersonResponse, error) {
	if request.Id != nil {
		s.Lock()
		defer s.Unlock()
		user, ok := s.users[*request.Id]
		if !ok {
			logger.Error("user not found")
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		logger.Info("user found")
		return &api.PersonResponse{Person: user}, nil
	}
	logger.Error("nil argument")
	return nil, status.Errorf(codes.InvalidArgument, "nil argument")
}
func (s *Server) AddPerson(_ context.Context, request *api.Person) (*api.PersonResponse, error) {
	if request.Id != nil && request.Name != "" {
		if _, ok := s.users[*request.Id]; ok {
			logger.Error("user already exists")
			return nil, status.Errorf(codes.AlreadyExists, "user already exists")
		}
		s.Lock()
		defer s.Unlock()
		s.users[*request.Id] = request
		logger.Info("user added")
		return &api.PersonResponse{Person: request}, nil
	}
	fmt.Println(request.Id)
	logger.Error("nil argument")
	return nil, status.Errorf(codes.InvalidArgument, "nil argument")
}

func StartGrpcServer() error {
	lis, err := net.Listen("tcp",
		fmt.Sprintf("%s:%d", "localhost", 8080))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to listen: %v", err))
		return err
	}
	s := grpc.NewServer()
	api.RegisterPersonServiceServer(s, &Server{users: make(map[int32]*api.Person), Mutex: sync.Mutex{}})
	logger.Info(fmt.Sprintf("Starting server on port %d", 8080))
	if err = s.Serve(lis); err != nil {
		logger.Error(fmt.Sprintf("Failed to start server: %v", err))
		return err
	}
	return nil
}
