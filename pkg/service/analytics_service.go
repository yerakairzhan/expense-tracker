package service

import (
	"context"
	"fmt"
	"time"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/cache"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

type AnalyticsService struct {
	txRepo analyticsRepository
	cache  *cache.AnalyticsCache
}

type analyticsRepository interface {
	LastMonthSummary(ctx context.Context, userID int64, start, end time.Time) (repository.AnalyticsSummaryRow, error)
	DailyProfit(ctx context.Context, userID int64, start, end time.Time) ([]repository.AnalyticsDailyProfitRow, error)
	LastMonthExpenseByCategory(ctx context.Context, userID int64, start, end time.Time) ([]repository.AnalyticsCategoryExpenseRow, error)
	MonthlyProfit(ctx context.Context, userID int64, startMonth, endMonth time.Time) ([]repository.AnalyticsMonthlyProfitRow, error)
	NetWorth(ctx context.Context, userID int64) (pgtype.Numeric, error)
}

func NewAnalyticsService(txRepo *repository.TransactionRepository, analyticsCache *cache.AnalyticsCache) *AnalyticsService {
	return &AnalyticsService{txRepo: txRepo, cache: analyticsCache}
}

func (s *AnalyticsService) LastMonthSummary(ctx context.Context, userID int64) (*models.AnalyticsSummary, *apperror.Error) {
	start, end := lastMonthRangeUTC(time.Now().UTC())

	cacheKey := cache.AnalyticsCacheKey(userID, "last-month-summary")
	var cached models.AnalyticsSummary
	if hit, _ := s.cache.Get(ctx, cacheKey, &cached); hit {
		return &cached, nil
	}

	row, err := s.txRepo.LastMonthSummary(ctx, userID, start, end)
	if err != nil {
		return nil, apperror.Internal("failed to load analytics summary")
	}
	result := &models.AnalyticsSummary{
		PeriodStart: start.Format("2006-01-02"),
		PeriodEnd:   end.Format("2006-01-02"),
		Income:      numericToString4(row.Income),
		Expense:     numericToString4(row.Expense),
		Profit:      numericToString4(row.Profit),
	}
	_ = s.cache.Set(ctx, cacheKey, result)
	return result, nil
}

func (s *AnalyticsService) DailyProfit(ctx context.Context, userID int64, query models.AnalyticsRangeQuery) ([]models.AnalyticsDailyPoint, *apperror.Error) {
	start, end, err := rangeFromQuery(query)
	if err != nil {
		return nil, apperror.Validation(err.Error())
	}

	cacheKey := cache.AnalyticsCacheKey(userID, "daily-profit", start.Format("2006-01-02"), end.Format("2006-01-02"))
	var cached []models.AnalyticsDailyPoint
	if hit, _ := s.cache.Get(ctx, cacheKey, &cached); hit {
		return cached, nil
	}

	rows, err := s.txRepo.DailyProfit(ctx, userID, start, end)
	if err != nil {
		return nil, apperror.Internal("failed to load analytics daily series")
	}
	out := make([]models.AnalyticsDailyPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.AnalyticsDailyPoint{
			Date:    row.Date.Format("2006-01-02"),
			Income:  numericToString4(row.Income),
			Expense: numericToString4(row.Expense),
			Profit:  numericToString4(row.Profit),
		})
	}
	_ = s.cache.Set(ctx, cacheKey, out)
	return out, nil
}

func (s *AnalyticsService) LastMonthExpenseByCategory(ctx context.Context, userID int64) ([]models.AnalyticsCategoryExpense, *apperror.Error) {
	start, end := lastMonthRangeUTC(time.Now().UTC())

	cacheKey := cache.AnalyticsCacheKey(userID, "last-month-categories")
	var cached []models.AnalyticsCategoryExpense
	if hit, _ := s.cache.Get(ctx, cacheKey, &cached); hit {
		return cached, nil
	}

	rows, err := s.txRepo.LastMonthExpenseByCategory(ctx, userID, start, end)
	if err != nil {
		return nil, apperror.Internal("failed to load analytics categories")
	}
	out := make([]models.AnalyticsCategoryExpense, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.AnalyticsCategoryExpense{
			Category: row.Category,
			Amount:   numericToString4(row.Amount),
		})
	}
	_ = s.cache.Set(ctx, cacheKey, out)
	return out, nil
}

func (s *AnalyticsService) MonthlyProfit(ctx context.Context, userID int64, query models.AnalyticsMonthlyProfitQuery) ([]models.AnalyticsMonthlyProfitPoint, *apperror.Error) {
	months := query.Months
	if months <= 0 {
		months = 6
	}
	now := time.Now().UTC()
	endMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startMonth := endMonth.AddDate(0, -(months-1), 0)

	cacheKey := cache.AnalyticsCacheKey(userID, "monthly-profit", fmt.Sprintf("%d", months))
	var cached []models.AnalyticsMonthlyProfitPoint
	if hit, _ := s.cache.Get(ctx, cacheKey, &cached); hit {
		return cached, nil
	}

	rows, err := s.txRepo.MonthlyProfit(ctx, userID, startMonth, endMonth)
	if err != nil {
		return nil, apperror.Internal("failed to load monthly profit")
	}
	out := make([]models.AnalyticsMonthlyProfitPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.AnalyticsMonthlyProfitPoint{
			Month:   row.Month.Format("2006-01"),
			Income:  numericToString4(row.Income),
			Expense: numericToString4(row.Expense),
			Profit:  numericToString4(row.Profit),
		})
	}
	_ = s.cache.Set(ctx, cacheKey, out)
	return out, nil
}

func (s *AnalyticsService) Summary(ctx context.Context, userID int64, query models.AnalyticsRangeQuery) (*models.AnalyticsSummary, *apperror.Error) {
	start, end, err := rangeFromQuery(query)
	if err != nil {
		return nil, apperror.Validation(err.Error())
	}

	cacheKey := cache.AnalyticsCacheKey(userID, "summary", start.Format("2006-01-02"), end.Format("2006-01-02"))
	var cached models.AnalyticsSummary
	if hit, _ := s.cache.Get(ctx, cacheKey, &cached); hit {
		return &cached, nil
	}

	row, err := s.txRepo.LastMonthSummary(ctx, userID, start, end)
	if err != nil {
		return nil, apperror.Internal("failed to load analytics summary")
	}
	result := &models.AnalyticsSummary{
		PeriodStart: start.Format("2006-01-02"),
		PeriodEnd:   end.Format("2006-01-02"),
		Income:      numericToString4(row.Income),
		Expense:     numericToString4(row.Expense),
		Profit:      numericToString4(row.Profit),
	}
	_ = s.cache.Set(ctx, cacheKey, result)
	return result, nil
}

func (s *AnalyticsService) ByCategory(ctx context.Context, userID int64, query models.AnalyticsRangeQuery) ([]models.AnalyticsCategoryExpense, *apperror.Error) {
	start, end, err := rangeFromQuery(query)
	if err != nil {
		return nil, apperror.Validation(err.Error())
	}

	cacheKey := cache.AnalyticsCacheKey(userID, "by-category", start.Format("2006-01-02"), end.Format("2006-01-02"))
	var cached []models.AnalyticsCategoryExpense
	if hit, _ := s.cache.Get(ctx, cacheKey, &cached); hit {
		return cached, nil
	}

	rows, err := s.txRepo.LastMonthExpenseByCategory(ctx, userID, start, end)
	if err != nil {
		return nil, apperror.Internal("failed to load analytics categories")
	}
	out := make([]models.AnalyticsCategoryExpense, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.AnalyticsCategoryExpense{
			Category: row.Category,
			Amount:   numericToString4(row.Amount),
		})
	}
	_ = s.cache.Set(ctx, cacheKey, out)
	return out, nil
}

func (s *AnalyticsService) Cashflow(ctx context.Context, userID int64, query models.AnalyticsRangeQuery) ([]models.AnalyticsDailyPoint, *apperror.Error) {
	start, end, err := rangeFromQuery(query)
	if err != nil {
		return nil, apperror.Validation(err.Error())
	}

	cacheKey := cache.AnalyticsCacheKey(userID, "cashflow", start.Format("2006-01-02"), end.Format("2006-01-02"))
	var cached []models.AnalyticsDailyPoint
	if hit, _ := s.cache.Get(ctx, cacheKey, &cached); hit {
		return cached, nil
	}

	rows, err := s.txRepo.DailyProfit(ctx, userID, start, end)
	if err != nil {
		return nil, apperror.Internal("failed to load cashflow")
	}
	out := make([]models.AnalyticsDailyPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.AnalyticsDailyPoint{
			Date:    row.Date.Format("2006-01-02"),
			Income:  numericToString4(row.Income),
			Expense: numericToString4(row.Expense),
			Profit:  numericToString4(row.Profit),
		})
	}
	_ = s.cache.Set(ctx, cacheKey, out)
	return out, nil
}

func (s *AnalyticsService) NetWorth(ctx context.Context, userID int64) (*models.AnalyticsNetWorth, *apperror.Error) {
	cacheKey := cache.AnalyticsCacheKey(userID, "net-worth")
	var cached models.AnalyticsNetWorth
	if hit, _ := s.cache.Get(ctx, cacheKey, &cached); hit {
		return &cached, nil
	}

	total, err := s.txRepo.NetWorth(ctx, userID)
	if err != nil {
		return nil, apperror.Internal("failed to load net worth")
	}
	result := &models.AnalyticsNetWorth{
		TotalBalance: numericToString4(total),
		AsOf:         time.Now().UTC().Format("2006-01-02"),
	}
	_ = s.cache.Set(ctx, cacheKey, result)
	return result, nil
}

func rangeFromQuery(query models.AnalyticsRangeQuery) (time.Time, time.Time, error) {
	if query.From == nil && query.To == nil {
		start, end := lastMonthRangeUTC(time.Now().UTC())
		return start, end, nil
	}

	if query.From == nil || query.To == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("both from and to are required when one is provided")
	}

	start, err := time.Parse("2006-01-02", *query.From)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid from date, expected YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", *query.To)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid to date, expected YYYY-MM-DD")
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("to must be greater than or equal to from")
	}
	if end.Sub(start) > 366*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("date range is too large (max 366 days)")
	}
	return start.UTC(), end.UTC(), nil
}

func lastMonthRangeUTC(now time.Time) (time.Time, time.Time) {
	firstOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstOfLastMonth := firstOfCurrentMonth.AddDate(0, -1, 0)
	lastOfLastMonth := firstOfCurrentMonth.AddDate(0, 0, -1)
	return firstOfLastMonth, lastOfLastMonth
}
