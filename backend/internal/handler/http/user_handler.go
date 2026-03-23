package http

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/aws"
	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/HeadTDev/fitchallenge/internal/domain/repositories"
	"github.com/HeadTDev/fitchallenge/internal/handler/http/middleware"
	"github.com/HeadTDev/fitchallenge/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	repo     repositories.UserRepository
	s3Client *aws.S3Client
}

func NewUserHandler(repo repositories.UserRepository, s3Client *aws.S3Client) *UserHandler {
	return &UserHandler{
		repo:     repo,
		s3Client: s3Client,
	}
}

// getUserID extracts and parses the userID from Gin context.
func (h *UserHandler) getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDStr := c.GetString(middleware.UserIDKey)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid user ID in context")
		return uuid.Nil, false
	}
	return userID, true
}

// MeHandler returns the logged-in user's basic info from context.
func (h *UserHandler) MeHandler(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	user, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "User not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch user")
		return
	}

	response.Success(c, http.StatusOK, user.ToResponse())
}

// GetProfile returns the full profile of the logged-in user.
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	user, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "User not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch user")
		return
	}

	response.Success(c, http.StatusOK, user.ToResponse())
}

type updateProfileRequest struct {
	DisplayName *string `json:"display_name" binding:"omitempty,min=2,max=50"`
	Bio         *string `json:"bio" binding:"omitempty,max=500"`
	Timezone    *string `json:"timezone" binding:"omitempty"`
}

// UpdateProfile updates the logged-in user's profile info.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	user, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "User not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch user")
		return
	}

	if req.DisplayName != nil {
		user.DisplayName = req.DisplayName
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}
	user.UpdatedAt = time.Now()

	if err := h.repo.Update(c.Request.Context(), user); err != nil {
		response.Error(c, http.StatusInternalServerError, "DB_ERROR", "Failed to update profile")
		return
	}

	response.Success(c, http.StatusOK, user.ToResponse())
}

// UploadAvatar handles avatar image upload to S3.
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "MISSING_FILE", "Avatar file is required")
		return
	}

	// Simple validation: check extension
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		response.Error(c, http.StatusBadRequest, "INVALID_TYPE", "Only JPG and PNG are allowed")
		return
	}

	// Open file
	f, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "FILE_ERROR", "Failed to open file")
		return
	}
	defer f.Close()

	// Upload to S3
	key := fmt.Sprintf("%s/%s%s", models.AvatarPrefix, userID.String(), ext)
	contentType := "image/jpeg"
	if ext == ".png" {
		contentType = "image/png"
	}

	avatarURL, err := h.s3Client.UploadFile(c.Request.Context(), models.AvatarBucket, key, f, contentType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "S3_ERROR", "Failed to upload to S3")
		return
	}

	// Update user record with new avatar URL
	user, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "User not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch user")
		return
	}

	user.AvatarURL = &avatarURL
	user.UpdatedAt = time.Now()

	if err := h.repo.Update(c.Request.Context(), user); err != nil {
		response.Error(c, http.StatusInternalServerError, "DB_ERROR", "Failed to update avatar URL")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"avatar_url": avatarURL,
	})
}
