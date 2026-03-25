package handler

import (
	"finance-tracker/pkg/apperror"
	"github.com/gin-gonic/gin"
)

func writeError(c *gin.Context, err *apperror.Error) {
	c.JSON(err.Status, gin.H{
		"error": gin.H{
			"code":    err.Code,
			"message": err.Message,
		},
	})
}
