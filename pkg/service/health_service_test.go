package service

import (
	"context"
	"errors"
	"testing"
)

func TestHealthService_Ready(t *testing.T) {
	t.Run("returns nil when ping succeeds", func(t *testing.T) {
		svc := &HealthService{
			pool: &fakePinger{pingFn: func(context.Context) error { return nil }},
		}

		// Act.
		err := svc.Ready(context.Background())

		// Assert.
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("maps ping failure", func(t *testing.T) {
		svc := &HealthService{
			pool: &fakePinger{pingFn: func(context.Context) error { return errors.New("down") }},
		}

		// Act.
		err := svc.Ready(context.Background())

		// Assert.
		if err == nil || err.Message != "database is not ready" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}
