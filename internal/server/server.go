package server

import (
	"context"
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	usersv1 "github.com/darkin-kat/store-api/gen/users/v1"
	"github.com/darkin-kat/user/internal/domain"
	"github.com/darkin-kat/user/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	usersv1.UnimplementedUserServiceServer
	repo repository.UserRepository
}

func NewServer(repo repository.UserRepository) *Server {
	return &Server{
		repo: repo,
	}
}

func (s *Server) Create(ctx context.Context, req *usersv1.CreateUserRequest) (*usersv1.CreateUserResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.GetFirstName() == "" {
		return nil, status.Error(codes.InvalidArgument, "first name is required")
	}
	if req.GetLastName() == "" {
		return nil, status.Error(codes.InvalidArgument, "last name is required")
	}
	if req.GetPassword() == "" || len(req.GetPassword()) < 8 {
		return nil, status.Error(codes.InvalidArgument, "password is required and must be at least 8 characters long")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}
	now := time.Now()
	
	user := &domain.User{
		ID:        uuid.New().String(),
		Email:     req.GetEmail(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		PasswordHash:  string(hashedPassword),
		CreatedAt: now,
		UpdatedAt: now,
	}
	res, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}
	return &usersv1.CreateUserResponse{
		User: toProtoUser(&res),
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *usersv1.GetUserRequest) (*usersv1.GetUserResponse, error) {
	var user domain.User
	var err error
	
	switch identifier := req.GetIdentifier().(type) {
		case *usersv1.GetUserRequest_Id:
			user, err = s.repo.GetByID(ctx, identifier.Id)
		case *usersv1.GetUserRequest_Email:
			user, err = s.repo.GetByEmail(ctx, identifier.Email)
		default:
			return nil, status.Error(codes.InvalidArgument, "user ID or email must be provided")
	}
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}
	return &usersv1.GetUserResponse{
		User: toProtoUser(&user),
	}, nil
}
	

func toProtoUser(user *domain.User) *usersv1.User {
	return &usersv1.User{
		Id:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}