package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tfenng/scaffold/internal/domain"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err

		var ae *domain.AppError
		if errors.As(err, &ae) {
			c.JSON(ae.HTTPStatus, gin.H{"code": ae.Code, "message": ae.Message})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"code": domain.CodeInternal, "message": "internal error"})
	}
}
