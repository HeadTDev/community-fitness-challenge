package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/HeadTDev/fitchallenge/internal/domain/repositories"
	"github.com/HeadTDev/fitchallenge/internal/pkg/jwt"
	"github.com/HeadTDev/fitchallenge/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	jwtManager *jwt.JWTManager
	repo       repositories.UserRepository
	appEnv     string
}

func NewAuthHandler(jwtManager *jwt.JWTManager, repo repositories.UserRepository, appEnv string) *AuthHandler {
	return &AuthHandler{
		jwtManager: jwtManager,
		repo:       repo,
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
	uID := uuid.New()
	userID := uID.String()
	role := string(models.RoleAdmin)

	// Persist user to DB so profile CRUD works
	user := &models.User{
		ID:          uID,
		Email:       fmt.Sprintf("dev-%s@fitchallenge.local", userID[:8]),
		DisplayName: stringPtr("Dev User"),
		Timezone:    "UTC",
		Role:        models.RoleAdmin,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.Create(c.Request.Context(), user); err != nil {
		response.Error(c, http.StatusInternalServerError, "DB_ERROR", "Failed to create dev user")
		return
	}

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

func stringPtr(s string) *string {
	return &s
}
