package service

import (
	"context"

	"finance-tracker/pkg/apperror"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthService struct {
	pool *pgxpool.Pool
}

func NewHealthService(pool *pgxpool.Pool) *HealthService {
	return &HealthService{pool: pool}
}

func (s *HealthService) Ready(ctx context.Context) *apperror.Error {
	if err := s.pool.Ping(ctx); err != nil {
		return apperror.Internal("database is not ready")
	}
	return nil
}
