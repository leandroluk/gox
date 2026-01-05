package cqrs

import (
	"context"
	"errors"
	"testing"

	"github.com/leandroluk/go/di"
)

// --- Mocks ---
type TestQuery struct{ ID int }
type TestResponse struct{ Name string }

type TestHandler struct{}

func (h *TestHandler) Handle(ctx context.Context, q TestQuery) (TestResponse, error) {
	if q.ID == 0 {
		return TestResponse{}, errors.New("not found")
	}
	return TestResponse{Name: "Result"}, nil
}

func TestCQRS_FullFlow(t *testing.T) {
	di.Reset()
	ctx := context.Background()

	// Registro do Handler
	RegisterQueryHandler[TestQuery, TestResponse, *TestHandler](func() *TestHandler {
		return &TestHandler{}
	})

	t.Run("Should execute successfully", func(t *testing.T) {
		res, err := ExecuteQuery[TestResponse](ctx, TestQuery{ID: 1})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res.Name != "Result" {
			t.Errorf("Expected 'Result', got %s", res.Name)
		}
	})

	t.Run("Should handle pointer coercion (Query as Pointer)", func(t *testing.T) {
		// Enviando *TestQuery para um handler que espera TestQuery
		res, err := ExecuteQuery[TestResponse](ctx, &TestQuery{ID: 1})
		if err != nil {
			t.Fatalf("Coercion failed: %v", err)
		}
		if res.Name != "Result" {
			t.Errorf("Expected 'Result', got %s", res.Name)
		}
	})

	t.Run("Should return error from handler", func(t *testing.T) {
		_, err := ExecuteQuery[TestResponse](ctx, TestQuery{ID: 0})
		if err == nil || err.Error() != "not found" {
			t.Errorf("Expected 'not found' error, got %v", err)
		}
	})
}
