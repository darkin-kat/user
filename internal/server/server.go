package server

import (
	"context"
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	usrv1 "github.com/darkin-kat/store-api/gen/users/v1"
	"github.com/darkin-kat/user/internal/domain"
	"github.com/darkin-kat/user/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	usrv1.UnimplementedUserServiceServer
	repo repository.UserRepository
}

func NewServer(repo repository.UserRepository) *Server {
	return &Server{
		repo: repo,
	}
}

func (s *Server) Create(ctx context.Context, req *usrv1.CreateUserRequest) (*usrv1.CreateUserResponse, error) {
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
	return &usrv1.CreateUserResponse{
		User: toProtoUser(&res),
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *usrv1.GetUserRequest) (*usrv1.GetUserResponse, error) {
	var user domain.User
	var err error
	
	switch identifier := req.GetIdentifier().(type) {
		case *usrv1.GetUserRequest_Id:
			user, err = s.repo.GetByID(ctx, identifier.Id)
		case *usrv1.GetUserRequest_Email:
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
	return &usrv1.GetUserResponse{
		User: toProtoUser(&user),
	}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *usrv1.UpdateUserRequest) (*usrv1.UpdateUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}
	user, err := s.repo.GetByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}
	if req.Email != nil {
		user.Email = req.GetEmail()
	}
	if req.FirstName != nil {
		user.FirstName = req.GetFirstName()
	}
	if req.LastName != nil {
		user.LastName = req.GetLastName()
	}
	if req.Password != nil {
		if len(req.GetPassword()) < 8 {
			return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters long")
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to hash password")
		}
		user.PasswordHash = string(hashedPassword)
	}
	user.UpdatedAt = time.Now()
	updatedUser, err := s.repo.Update(ctx, &user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update user")
	}
	return &usrv1.UpdateUserResponse{
		User: toProtoUser(&updatedUser),
	}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *usrv1.DeleteUserRequest) (*usrv1.DeleteUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}
	err := s.repo.Delete(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete user")
	}
	return &usrv1.DeleteUserResponse{}, nil
}

func (s *Server) ListUsers(ctx context.Context, req *usrv1.ListUsersRequest) (*usrv1.ListUsersResponse, error) {
	if req.GetLimit() <= 0 {
		req.Limit = 10
	}
	if req.GetOffset() <= 0 {
		req.Offset = 0
	}
	users, err := s.repo.List(ctx, uint(req.GetLimit()), uint(req.GetOffset()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list users")
	}
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to count users")
	}
	protoUsers := make([]*usrv1.User, len(users))
	for i := range users {
		protoUsers[i] = toProtoUser(&users[i])
	}
	return &usrv1.ListUsersResponse{
		Users: protoUsers,
		Total: int32(total),
	}, nil
}

func toProtoUser(user *domain.User) *usrv1.User {
	return &usrv1.User{
		Id:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}