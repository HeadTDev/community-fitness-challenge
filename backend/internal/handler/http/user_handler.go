package http

import (
	"net/http"

	"github.com/HeadTDev/fitchallenge/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// MeHandler visszaadja a bejelentkezett felhasználó adatait a kontextusból.
func MeHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("role")

	response.Success(c, http.StatusOK, gin.H{
		"user_id": userID,
		"role":    role,
	})
}
