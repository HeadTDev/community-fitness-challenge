package http

import (
	"net/http"

	"github.com/HeadTDev/fitchallenge/internal/pkg/jwt"
	"github.com/HeadTDev/fitchallenge/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	jwtManager *jwt.JWTManager
	appEnv     string
}

func NewAuthHandler(jwtManager *jwt.JWTManager, appEnv string) *AuthHandler {
	return &AuthHandler{
		jwtManager: jwtManager,
		appEnv:     appEnv,
	}
}

// RegisterDev egy ideiglenes végpont a fejlesztéshez, ami azonnal ad egy tokent.
func (h *AuthHandler) RegisterDev(c *gin.Context) {
	// Biztonsági check: Produkcióban tilos!
	if h.appEnv == "production" {
		response.Error(c, http.StatusForbidden, "FORBIDDEN", "Development endpoint is disabled in production")
		return
	}

	// Teszt user generálása
	userID := uuid.New().String()
	role := "user"

	accessToken, err := h.jwtManager.GenerateAccessToken(userID, role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "TOKEN_GEN_FAILED", "Failed to generate access token")
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "TOKEN_GEN_FAILED", "Failed to generate refresh token")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"user_id":       userID,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// RefreshToken megújítja az access tokent egy érvényes refresh token alapján.
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	claims, err := h.jwtManager.ValidateToken(input.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", err.Error())
		return
	}

	// Új access token generálása (a role-t itt most egyszerűség kedvéért visszaadjuk)
	newAccessToken, err := h.jwtManager.GenerateAccessToken(claims.UserID, claims.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "TOKEN_GEN_FAILED", "Failed to generate new access token")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"access_token": newAccessToken,
	})
}
