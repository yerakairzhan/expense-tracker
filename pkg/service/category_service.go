package service

import (
	"context"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"
)

type CategoryService struct {
	repo categoryRepository
}

type categoryRepository interface {
	ListForUser(ctx context.Context, userID int64) ([]repository.CategoryRow, error)
	Create(ctx context.Context, userID int64, name, catType string, color, icon *string) (repository.CategoryRow, error)
	GetByIDForUser(ctx context.Context, id, userID int64) (repository.CategoryRow, error)
	Update(ctx context.Context, id, userID int64, name, color, icon *string) (repository.CategoryRow, error)
	Delete(ctx context.Context, id, userID int64) (int64, error)
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List(ctx context.Context, userID int64) ([]models.Category, *apperror.Error) {
	rows, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, apperror.Internal("failed to list categories")
	}
	out := make([]models.Category, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapCategory(r))
	}
	return out, nil
}

func (s *CategoryService) Create(ctx context.Context, userID int64, req models.CreateCategoryRequest) (*models.Category, *apperror.Error) {
	row, err := s.repo.Create(ctx, userID, req.Name, req.Type, req.Color, req.Icon)
	if err != nil {
		return nil, apperror.Internal("failed to create category")
	}
	out := mapCategory(row)
	return &out, nil
}

func (s *CategoryService) Update(ctx context.Context, userID, id int64, req models.UpdateCategoryRequest) (*models.Category, *apperror.Error) {
	if req.Name == nil && req.Color == nil && req.Icon == nil {
		return nil, apperror.Validation("at least one field is required")
	}
	row, err := s.repo.Update(ctx, id, userID, req.Name, req.Color, req.Icon)
	if err != nil {
		return nil, apperror.NotFound("category not found or is a system category")
	}
	out := mapCategory(row)
	return &out, nil
}

func (s *CategoryService) Delete(ctx context.Context, userID, id int64) *apperror.Error {
	affected, err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		return apperror.Internal("failed to delete category")
	}
	if affected == 0 {
		return apperror.NotFound("category not found or is a system category")
	}
	return nil
}
