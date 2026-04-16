package handler

import (
	"context"
	"net/http"
	"strconv"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/middleware"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/service"

	"github.com/gin-gonic/gin"
)

type recurringService interface {
	List(ctx context.Context, userID int64) ([]models.RecurringPayment, *apperror.Error)
	Create(ctx context.Context, userID int64, req models.CreateRecurringRequest) (*models.RecurringPayment, *apperror.Error)
	Update(ctx context.Context, userID, id int64, req models.UpdateRecurringRequest) (*models.RecurringPayment, *apperror.Error)
	Delete(ctx context.Context, userID, id int64) *apperror.Error
}

type RecurringHandler struct {
	svc recurringService
}

func NewRecurringHandler(svc *service.RecurringService) *RecurringHandler {
	return &RecurringHandler{svc: svc}
}

func (h *RecurringHandler) List(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	out, appErr := h.svc.List(c.Request.Context(), userID)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *RecurringHandler) Create(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	var req models.CreateRecurringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.svc.Create(c.Request.Context(), userID, req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *RecurringHandler) Update(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, apperror.Validation("invalid recurring payment id"))
		return
	}
	var req models.UpdateRecurringRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.svc.Update(c.Request.Context(), userID, id, req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *RecurringHandler) Delete(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, apperror.Validation("invalid recurring payment id"))
		return
	}
	if appErr := h.svc.Delete(c.Request.Context(), userID, id); appErr != nil {
		writeError(c, appErr)
		return
	}
	c.Status(http.StatusNoContent)
}
