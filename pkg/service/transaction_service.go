package service

import (
	"context"
	"strings"

	sqlc "finance-tracker/db/queries"
	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

// AnalyticsCacheInvalidator is satisfied by *cache.AnalyticsCache.
// Keeping it as an interface avoids a direct import of the cache package here.
type AnalyticsCacheInvalidator interface {
	InvalidateUser(ctx context.Context, userID int64) error
}

type TransactionService struct {
	txRepo         transactionRepository
	analyticsCache AnalyticsCacheInvalidator
}

type transactionRepository interface {
	ListForUser(ctx context.Context, params sqlc.ListTransactionsForUserParams) ([]sqlc.Transaction, error)
	CreateForUser(ctx context.Context, userID int64, params sqlc.CreateTransactionParams) (sqlc.Transaction, error)
	GetByIDForUser(ctx context.Context, txID, userID int64) (sqlc.Transaction, error)
	UpdateForUser(ctx context.Context, userID, txID int64, params sqlc.UpdateTransactionByIDForUserParams) (sqlc.Transaction, error)
	SoftDeleteForUser(ctx context.Context, userID, txID int64) error
}

func NewTransactionService(txRepo *repository.TransactionRepository, analyticsCache AnalyticsCacheInvalidator) *TransactionService {
	return &TransactionService{txRepo: txRepo, analyticsCache: analyticsCache}
}

func (s *TransactionService) List(ctx context.Context, userID int64, query models.ListTransactionsQuery) ([]models.Transaction, *apperror.Error) {
	offset := (query.Page - 1) * query.Limit

	params := sqlc.ListTransactionsForUserParams{
		UserID:     userID,
		OffsetRows: int32(offset),
		LimitRows:  int32(query.Limit),
	}

	if query.AccountID != nil {
		params.AccountID = pgtype.Int8{Int64: *query.AccountID, Valid: true}
	}
	if query.CategoryID != nil {
		params.CategoryID = pgtype.Int8{Int64: *query.CategoryID, Valid: true}
	}
	if query.Type != nil {
		params.Type = pgtype.Text{String: *query.Type, Valid: true}
	}
	if query.From != nil {
		from, err := dateFromString(*query.From)
		if err != nil {
			return nil, apperror.Validation(err.Error())
		}
		params.FromDate = from
	}
	if query.To != nil {
		to, err := dateFromString(*query.To)
		if err != nil {
			return nil, apperror.Validation(err.Error())
		}
		params.ToDate = to
	}

	rows, err := s.txRepo.ListForUser(ctx, params)
	if err != nil {
		return nil, apperror.Internal("failed to list transactions")
	}
	out := make([]models.Transaction, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapTransaction(row))
	}
	return out, nil
}

func (s *TransactionService) Create(ctx context.Context, userID int64, req models.CreateTransactionRequest) (*models.Transaction, *apperror.Error) {
	if !isPositiveDecimal(req.Amount) {
		return nil, apperror.Validation("amount must be a positive decimal with up to 4 fraction digits")
	}
	date, err := dateFromString(req.TransactedAt)
	if err != nil {
		return nil, apperror.Validation(err.Error())
	}
	amount, err := stringToNumeric(req.Amount)
	if err != nil {
		return nil, apperror.Validation(err.Error())
	}

	row, err := s.txRepo.CreateForUser(ctx, userID, sqlc.CreateTransactionParams{
		AccountID:    req.AccountID,
		CategoryID:   int8FromPtr(req.CategoryID),
		Amount:       amount,
		Currency:     req.Currency,
		Type:         req.Type,
		Description:  req.Description,
		Notes:        textFromPtr(req.Notes),
		TransactedAt: date,
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return nil, apperror.NotFound("account not found")
		}
		return nil, apperror.Internal("failed to create transaction")
	}
	if s.analyticsCache != nil {
		_ = s.analyticsCache.InvalidateUser(ctx, userID)
	}
	out := mapTransaction(row)
	return &out, nil
}

func (s *TransactionService) GetByID(ctx context.Context, userID, txID int64) (*models.Transaction, *apperror.Error) {
	row, err := s.txRepo.GetByIDForUser(ctx, txID, userID)
	if err != nil {
		return nil, apperror.NotFound("transaction not found")
	}
	out := mapTransaction(row)
	return &out, nil
}

func (s *TransactionService) Update(ctx context.Context, userID, txID int64, req models.UpdateTransactionRequest) (*models.Transaction, *apperror.Error) {
	if req.Amount == nil && req.CategoryID == nil && req.Notes == nil {
		return nil, apperror.Validation("at least one field is required")
	}

	amount := pgtype.Numeric{}
	if req.Amount != nil {
		if !isPositiveDecimal(*req.Amount) {
			return nil, apperror.Validation("amount must be a positive decimal with up to 4 fraction digits")
		}
		numeric, err := stringToNumeric(*req.Amount)
		if err != nil {
			return nil, apperror.Validation(err.Error())
		}
		amount = numeric
	}

	row, err := s.txRepo.UpdateForUser(ctx, userID, txID, sqlc.UpdateTransactionByIDForUserParams{
		Amount:     amount,
		CategoryID: int8FromPtr(req.CategoryID),
		Notes:      textFromPtr(req.Notes),
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return nil, apperror.NotFound("transaction not found")
		}
		return nil, apperror.Internal("failed to update transaction")
	}
	if s.analyticsCache != nil {
		_ = s.analyticsCache.InvalidateUser(ctx, userID)
	}
	out := mapTransaction(row)
	return &out, nil
}

func (s *TransactionService) Delete(ctx context.Context, userID, txID int64) *apperror.Error {
	if err := s.txRepo.SoftDeleteForUser(ctx, userID, txID); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return apperror.NotFound("transaction not found")
		}
		return apperror.Internal("failed to delete transaction")
	}
	if s.analyticsCache != nil {
		_ = s.analyticsCache.InvalidateUser(ctx, userID)
	}
	return nil
}
