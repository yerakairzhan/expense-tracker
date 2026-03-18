package handler

import (
	"net/http"

	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"
	"github.com/gin-gonic/gin"
)

// AccountHandler handles account-related HTTP requests
type AccountHandler struct {
	repo *repository.AccountRepository
}

// NewAccountHandler creates a new AccountHandler
func NewAccountHandler(repo *repository.AccountRepository) *AccountHandler {
	return &AccountHandler{repo: repo}
}

// Create godoc
// @Summary Create account
// @Description Create a new financial account
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body CreateAccountRequest true "Create account payload"
// @Success 201 {object} Account
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /accounts [post]
func (h *AccountHandler) Create(c *gin.Context) {
	var req models.CreateAccountRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	account, err := h.repo.CreateAccount(
		c.Request.Context(),
		req.UserID,
		req.AccountType,
		req.Balance,
		req.Currency,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}
	
	c.JSON(http.StatusCreated, account)
}
