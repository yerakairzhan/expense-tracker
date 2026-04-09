package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"
)

func TestAnalyticsService_DailyProfit(t *testing.T) {
	from := "2024-01-01"
	to := "2024-01-03"
	svc := &AnalyticsService{
		txRepo: &fakeAnalyticsRepo{
			dailyProfitFn: func(_ context.Context, userID int64, start, end time.Time) ([]repository.AnalyticsDailyProfitRow, error) {
				if userID != 42 || start.Format("2006-01-02") != from || end.Format("2006-01-02") != to {
					t.Fatalf("unexpected range: %d %s %s", userID, start, end)
				}
				return []repository.AnalyticsDailyProfitRow{{
					Date:    time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
					Income:  mustNumeric("12.0000"),
					Expense: mustNumeric("5.0000"),
					Profit:  mustNumeric("7.0000"),
				}}, nil
			},
		},
	}

	// Act.
	got, err := svc.DailyProfit(context.Background(), 42, models.AnalyticsRangeQuery{From: &from, To: &to})

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}
	if len(got) != 1 || got[0].Profit != "7.0000" {
		t.Fatalf("unexpected points: %#v", got)
	}
}

func TestAnalyticsService_DailyProfitValidation(t *testing.T) {
	to := "2024-01-03"
	svc := &AnalyticsService{txRepo: &fakeAnalyticsRepo{}}

	// Act.
	got, err := svc.DailyProfit(context.Background(), 42, models.AnalyticsRangeQuery{To: &to})

	// Assert.
	if got != nil {
		t.Fatal("expected nil points")
	}
	if err == nil || err.Message != "both from and to are required when one is provided" {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func TestAnalyticsService_MonthlyProfitAndSummary(t *testing.T) {
	t.Run("monthly profit defaults to six months", func(t *testing.T) {
		svc := &AnalyticsService{
			txRepo: &fakeAnalyticsRepo{
				monthlyProfitFn: func(_ context.Context, userID int64, startMonth, endMonth time.Time) ([]repository.AnalyticsMonthlyProfitRow, error) {
					if userID != 42 || endMonth.Sub(startMonth) < 5*28*24*time.Hour {
						t.Fatalf("unexpected month range: %s %s", startMonth, endMonth)
					}
					return []repository.AnalyticsMonthlyProfitRow{{
						Month:   time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
						Income:  mustNumeric("10.0000"),
						Expense: mustNumeric("4.0000"),
						Profit:  mustNumeric("6.0000"),
					}}, nil
				},
			},
		}

		// Act.
		got, err := svc.MonthlyProfit(context.Background(), 42, models.AnalyticsMonthlyProfitQuery{})

		// Assert.
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
		if len(got) != 1 || got[0].Month != "2024-01" {
			t.Fatalf("unexpected points: %#v", got)
		}
	})

	t.Run("summary maps repo error", func(t *testing.T) {
		svc := &AnalyticsService{
			txRepo: &fakeAnalyticsRepo{
				lastMonthSummaryFn: func(context.Context, int64, time.Time, time.Time) (repository.AnalyticsSummaryRow, error) {
					return repository.AnalyticsSummaryRow{}, errors.New("boom")
				},
			},
		}

		// Act.
		got, err := svc.LastMonthSummary(context.Background(), 42)

		// Assert.
		if got != nil {
			t.Fatal("expected nil summary")
		}
		if err == nil || err.Message != "failed to load analytics summary" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestRangeFromQuery(t *testing.T) {
	t.Run("rejects reversed range", func(t *testing.T) {
		from := "2024-01-03"
		to := "2024-01-01"

		// Act.
		_, _, err := rangeFromQuery(models.AnalyticsRangeQuery{From: &from, To: &to})

		// Assert.
		if err == nil || err.Error() != "to must be greater than or equal to from" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects too-large range", func(t *testing.T) {
		from := "2024-01-01"
		to := "2025-01-03"

		// Act.
		_, _, err := rangeFromQuery(models.AnalyticsRangeQuery{From: &from, To: &to})

		// Assert.
		if err == nil || err.Error() != "date range is too large (max 366 days)" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns explicit range", func(t *testing.T) {
		from := "2024-01-01"
		to := "2024-01-02"

		// Act.
		start, end, err := rangeFromQuery(models.AnalyticsRangeQuery{From: &from, To: &to})

		// Assert.
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if start.Format("2006-01-02") != from || end.Format("2006-01-02") != to {
			t.Fatalf("unexpected range: %s %s", start, end)
		}
	})
}
