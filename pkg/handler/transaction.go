package handler

import (
	"net/http"
	"strconv"

	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"

	"github.com/gin-gonic/gin"
)

// TransactionHandler handles transaction-related HTTP requests
type TransactionHandler struct {
	repo *repository.TransactionRepository
}

// NewTransactionHandler creates a new TransactionHandler
func NewTransactionHandler(repo *repository.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{repo: repo}
}

// Create godoc
// @Summary Create transaction
// @Description Create a new transaction for an account
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body CreateTransactionRequest true "Create transaction payload"
// @Success 201 {object} Transaction
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /transactions [post]
func (h *TransactionHandler) Create(c *gin.Context) {
	var req models.CreateTransactionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.repo.CreateTransaction(
		c.Request.Context(),
		req.AccountID,
		req.Amount,
		req.Description,
		req.Type,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, tx)
}

// List returns a list of transactions (either all or filtered by account_id)
// GET /transactions?account_id
// List godoc
// @Summary List transactions
// @Description List transactions; optionally filter by account_id
// @Tags transactions
// @Produce json
// @Param account_id query int false "Account ID filter"
// @Param limit query int false "Limit (default 50)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {array} Transaction
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /transactions [get]
func (h *TransactionHandler) List(c *gin.Context) {
	accountIDStr := c.Query("account_id")

	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		parsedLimit, err := strconv.Atoi(l)
		if err != nil || parsedLimit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer"})
			return
		}
		limit = parsedLimit
	}

	if o := c.Query("offset"); o != "" {
		parsedOffset, err := strconv.Atoi(o)
		if err != nil || parsedOffset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be non-negative"})
			return
		}
		offset = parsedOffset
	}

	var (
		transactions []models.Transaction
		err          error
	)

	// ✅ Conditional logic
	if accountIDStr != "" {
		accountID, err := strconv.Atoi(accountIDStr)
		if err != nil || accountID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account_id"})
			return
		}

		transactions, err = h.repo.ListTransactionsByAccountID(
			c.Request.Context(),
			accountID,
			limit,
			offset,
		)
	} else {
		transactions, err = h.repo.List(
			c.Request.Context(),
			limit,
			offset,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// GetByID godoc
// @Summary Get transaction by ID
// @Tags transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} Transaction
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /transactions/{id} [get]
func (h *TransactionHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	tx, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// Delete godoc
// @Summary Delete transaction
// @Tags transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /transactions/{id} [delete]
func (h *TransactionHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "deleted"})
}

// GetByAccount godoc
// @Summary List transactions for an account
// @Tags transactions
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {array} Transaction
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /accounts/{id}/transactions [get]
func (h *TransactionHandler) GetByAccount(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	limit := 50
	offset := 0

	txs, err := h.repo.ListTransactionsByAccountID(
		c.Request.Context(),
		id,
		limit,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch"})
		return
	}

	c.JSON(http.StatusOK, txs)
}

// Search godoc
// @Summary Search transactions
// @Tags transactions
// @Produce json
// @Param q query string true "Search keyword"
// @Success 200 {array} Transaction
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /transactions/search [get]
func (h *TransactionHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	txs, err := h.repo.Search(c.Request.Context(), query, 50, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	c.JSON(http.StatusOK, txs)
}

// Export godoc
// @Summary Export transactions
// @Description Download transactions as JSON (Content-Disposition attachment)
// @Tags transactions
// @Produce json
// @Success 200 {array} Transaction
// @Failure 500 {object} ErrorResponse
// @Header 200 {string} Content-Disposition "attachment; filename=transactions.json"
// @Router /transactions/export [get]
func (h *TransactionHandler) Export(c *gin.Context) {
	txs, err := h.repo.List(c.Request.Context(), 1000, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "export failed"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=transactions.json")
	c.JSON(http.StatusOK, txs)
}
