package http

import (
	"net/http"

	"github.com/HeadTDev/fitchallenge/internal/handler/http/middleware"
	"github.com/HeadTDev/fitchallenge/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// MeHandler visszaadja a bejelentkezett felhasználó adatait a kontextusból.
func MeHandler(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	role := c.GetString(middleware.RoleKey)

	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"user_id": userID,
		"role":    role,
	})
}
