package handler

import "finance-tracker/pkg/models"

// ErrorResponse matches the {"error": "..."} JSON shape used by handlers.
type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type BalanceResponse struct {
	Balance string `json:"balance"`
}

// These aliases let swagger annotations refer to request/response types without
// importing pkg/models in every handler file (avoids unused imports).
type (
	User                     = models.User
	Account                  = models.Account
	Transaction              = models.Transaction
	RegisterRequest          = models.RegisterRequest
	UpdateUserRequest        = models.UpdateUserRequest
	CreateAccountRequest     = models.CreateAccountRequest
	CreateTransactionRequest = models.CreateTransactionRequest
	ListTransactionsQuery    = models.ListTransactionsQuery
)
