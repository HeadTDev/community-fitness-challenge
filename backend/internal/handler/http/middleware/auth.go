package middleware

import (
	"net/http"
	"strings"

	"github.com/HeadTDev/fitchallenge/internal/pkg/jwt"
	"github.com/HeadTDev/fitchallenge/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware ellenőrzi a JWT tokent az Authorization fejlécben.
func AuthMiddleware(jwtManager *jwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "AUTH_REQUIRED", "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "INVALID_AUTH_FORMAT", "Authorization header format must be Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", err.Error())
			c.Abort()
			return
		}

		// Elmentjük a claim-eket a kontextusba a későbbi használathoz
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}
