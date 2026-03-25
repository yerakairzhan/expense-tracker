package service

import (
	"context"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Me(ctx context.Context, userID int64) (*models.User, *apperror.Error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.NotFound("user not found")
	}
	out := mapUser(user)
	return &out, nil
}

func (s *UserService) UpdateMe(ctx context.Context, userID int64, req models.UpdateMeRequest) (*models.User, *apperror.Error) {
	if req.Name == nil && req.Currency == nil {
		return nil, apperror.Validation("at least one field is required")
	}
	user, err := s.users.UpdateProfile(ctx, userID, req.Name, req.Currency)
	if err != nil {
		return nil, apperror.Internal("failed to update user")
	}
	out := mapUser(user)
	return &out, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID int64, req models.ChangePasswordRequest) *apperror.Error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return apperror.NotFound("user not found")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)) != nil {
		return apperror.Unauthorized("current password is incorrect")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperror.Internal("failed to hash password")
	}
	if _, err = s.users.UpdatePassword(ctx, userID, string(newHash)); err != nil {
		return apperror.Internal("failed to update password")
	}
	return nil
}
