package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/models"
	"github.com/gin-gonic/gin"
)

type analyticsServiceSpy struct {
	lastMonthSummaryFn           func(ctx context.Context, userID int64) (*models.AnalyticsSummary, *apperror.Error)
	dailyProfitFn                func(ctx context.Context, userID int64, query models.AnalyticsRangeQuery) ([]models.AnalyticsDailyPoint, *apperror.Error)
	lastMonthExpenseByCategoryFn func(ctx context.Context, userID int64) ([]models.AnalyticsCategoryExpense, *apperror.Error)
	monthlyProfitFn              func(ctx context.Context, userID int64, query models.AnalyticsMonthlyProfitQuery) ([]models.AnalyticsMonthlyProfitPoint, *apperror.Error)
}

func (s *analyticsServiceSpy) LastMonthSummary(ctx context.Context, userID int64) (*models.AnalyticsSummary, *apperror.Error) {
	return s.lastMonthSummaryFn(ctx, userID)
}

func (s *analyticsServiceSpy) DailyProfit(ctx context.Context, userID int64, query models.AnalyticsRangeQuery) ([]models.AnalyticsDailyPoint, *apperror.Error) {
	return s.dailyProfitFn(ctx, userID, query)
}

func (s *analyticsServiceSpy) LastMonthExpenseByCategory(ctx context.Context, userID int64) ([]models.AnalyticsCategoryExpense, *apperror.Error) {
	return s.lastMonthExpenseByCategoryFn(ctx, userID)
}

func (s *analyticsServiceSpy) MonthlyProfit(ctx context.Context, userID int64, query models.AnalyticsMonthlyProfitQuery) ([]models.AnalyticsMonthlyProfitPoint, *apperror.Error) {
	return s.monthlyProfitFn(ctx, userID, query)
}

func TestAnalyticsHandler_LastMonthSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 401 without user context", func(t *testing.T) {
		router := gin.New()
		router.GET("/analytics/summary", (&AnalyticsHandler{analyticsService: &analyticsServiceSpy{}}).LastMonthSummary)

		// Act.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/analytics/summary", nil)
		router.ServeHTTP(rec, req)

		// Assert.
		assertStatus(t, rec, http.StatusUnauthorized)
		assertErrorEnvelope(t, rec, "UNAUTHORIZED", "invalid token context")
	})

	t.Run("returns 200 on success", func(t *testing.T) {
		service := &analyticsServiceSpy{
			lastMonthSummaryFn: func(_ context.Context, userID int64) (*models.AnalyticsSummary, *apperror.Error) {
				if userID != 42 {
					t.Fatalf("unexpected user id: %d", userID)
				}
				return &models.AnalyticsSummary{Profit: "7.0000"}, nil
			},
		}
		router := gin.New()
		router.GET("/analytics/summary", withUserID(42, (&AnalyticsHandler{analyticsService: service}).LastMonthSummary))

		// Act.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/analytics/summary", nil)
		router.ServeHTTP(rec, req)

		// Assert.
		assertStatus(t, rec, http.StatusOK)
	})
}

func TestAnalyticsHandler_DailyProfitAndMonthlyProfit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("daily profit binds query and returns 200", func(t *testing.T) {
		service := &analyticsServiceSpy{
			dailyProfitFn: func(_ context.Context, userID int64, query models.AnalyticsRangeQuery) ([]models.AnalyticsDailyPoint, *apperror.Error) {
				if userID != 42 || query.From == nil || *query.From != "2024-01-01" {
					t.Fatalf("unexpected query: %d %+v", userID, query)
				}
				return []models.AnalyticsDailyPoint{{Date: "2024-01-01", Profit: "7.0000"}}, nil
			},
		}
		router := gin.New()
		router.GET("/analytics/daily-profit", withUserID(42, (&AnalyticsHandler{analyticsService: service}).DailyProfit))

		// Act.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/analytics/daily-profit?from=2024-01-01&to=2024-01-02", nil)
		router.ServeHTTP(rec, req)

		// Assert.
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("monthly profit maps service error", func(t *testing.T) {
		service := &analyticsServiceSpy{
			monthlyProfitFn: func(context.Context, int64, models.AnalyticsMonthlyProfitQuery) ([]models.AnalyticsMonthlyProfitPoint, *apperror.Error) {
				return nil, apperror.Internal("failed to load monthly profit")
			},
		}
		router := gin.New()
		router.GET("/analytics/monthly-profit", withUserID(42, (&AnalyticsHandler{analyticsService: service}).MonthlyProfit))

		// Act.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/analytics/monthly-profit?months=6", nil)
		router.ServeHTTP(rec, req)

		// Assert.
		assertStatus(t, rec, http.StatusInternalServerError)
		assertErrorEnvelope(t, rec, "INTERNAL_ERROR", "failed to load monthly profit")
	})
}

func TestAnalyticsHandler_LastMonthExpenseByCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &analyticsServiceSpy{
		lastMonthExpenseByCategoryFn: func(_ context.Context, userID int64) ([]models.AnalyticsCategoryExpense, *apperror.Error) {
			if userID != 42 {
				t.Fatalf("unexpected user id: %d", userID)
			}
			return []models.AnalyticsCategoryExpense{{Category: "Food", Amount: "5.0000"}}, nil
		},
	}
	router := gin.New()
	router.GET("/analytics/expense-categories", withUserID(42, (&AnalyticsHandler{analyticsService: service}).LastMonthExpenseByCategory))

	// Act.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics/expense-categories", nil)
	router.ServeHTTP(rec, req)

	// Assert.
	assertStatus(t, rec, http.StatusOK)
}
