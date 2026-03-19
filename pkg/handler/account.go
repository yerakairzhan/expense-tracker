package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"
)

type AccountHandler struct {
	repo *repository.AccountRepository
}

func NewAccountHandler(r *repository.AccountRepository) *AccountHandler {
	return &AccountHandler{repo: r}
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

	account, err := h.repo.Create(
		c.Request.Context(),
		req.UserID,
		req.AccountType,
		req.Currency,
		req.Balance,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}

// GetByID godoc
// @Summary Get account by ID
// @Tags accounts
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} Account
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /accounts/{id} [get]
func (h *AccountHandler) GetByID(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	account, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

// List godoc
// @Summary List accounts
// @Tags accounts
// @Produce json
// @Success 200 {array} Account
// @Failure 500 {object} ErrorResponse
// @Router /accounts [get]
func (h *AccountHandler) List(c *gin.Context) {

	accounts, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// GetUserAccounts godoc
// @Summary List accounts for a user
// @Tags accounts
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {array} Account
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/accounts [get]
func (h *AccountHandler) GetUserAccounts(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	accounts, err := h.repo.GetByUserID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// Delete godoc
// @Summary Delete account
// @Tags accounts
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /accounts/{id} [delete]
func (h *AccountHandler) Delete(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	err = h.repo.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "account deleted"})
}

// GetBalance godoc
// @Summary Get account balance
// @Tags accounts
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} BalanceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /accounts/{id}/balance [get]
func (h *AccountHandler) GetBalance(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	balance, err := h.repo.GetBalance(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, BalanceResponse{Balance: balance})
}
