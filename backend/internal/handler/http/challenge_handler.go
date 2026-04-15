package http

import (
	"errors"
	"net/http"
	"path/filepath"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/HeadTDev/fitchallenge/internal/domain/services"
	"github.com/HeadTDev/fitchallenge/internal/handler/http/middleware"
	"github.com/HeadTDev/fitchallenge/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChallengeHandler struct {
	service            services.ChallengeService
	logService         services.LogService
	leaderboardService services.LeaderboardService
}

func NewChallengeHandler(service services.ChallengeService, logService services.LogService, leaderboardService services.LeaderboardService) *ChallengeHandler {
	return &ChallengeHandler{
		service:            service,
		logService:         logService,
		leaderboardService: leaderboardService,
	}
}

func (h *ChallengeHandler) getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDStr := c.GetString(middleware.UserIDKey)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid user ID in token")
		return uuid.Nil, false
	}
	return userID, true
}

type createChallengeRequest struct {
	Title           string               `json:"title" binding:"required,min=3,max=100"`
	Description     *string              `json:"description" binding:"omitempty,max=1000"`
	StartDate       time.Time            `json:"start_date" binding:"required"`
	EndDate         time.Time            `json:"end_date" binding:"required,gtfield=StartDate"`
	Type            models.ChallengeType `json:"type" binding:"required"`
	Goal            int                  `json:"goal" binding:"required,min=1"`
	MaxParticipants int                  `json:"max_participants" binding:"omitempty,min=0"`
}

type prizeRequest struct {
	Title        string  `json:"title" binding:"required,min=3,max=100"`
	Description  *string `json:"description" binding:"omitempty,max=500"`
	ImageURL     *string `json:"image_url" binding:"omitempty,url"`
	RankRequired int     `json:"rank_required" binding:"required,min=1"`
}

type submitDailyLogRequest struct {
	LogDate           *time.Time `json:"log_date" binding:"omitempty"`
	Steps             int        `json:"steps" binding:"required,min=0"`
	Calories          int        `json:"calories" binding:"required,min=0"`
	ActiveMinutes     int        `json:"active_minutes" binding:"required,min=0"`
	HealthKitDataHash *string    `json:"healthkit_data_hash" binding:"omitempty,max=128"`
	SourceBundleIDs   []string   `json:"source_bundle_ids" binding:"omitempty"`
}

func (h *ChallengeHandler) CreateChallenge(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var req createChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	challenge := &models.Challenge{
		Title:           req.Title,
		Description:     req.Description,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		Type:            req.Type,
		Goal:            req.Goal,
		MaxParticipants: req.MaxParticipants,
	}

	err := h.service.CreateChallenge(c.Request.Context(), userID, challenge)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "Only creators or admins can create challenges")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create challenge")
		return
	}

	response.Success(c, http.StatusCreated, challenge.ToResponse())
}

func (h *ChallengeHandler) AddPrize(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	idParam := c.Param("id")
	challengeID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	var req prizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	prize := &models.Prize{
		Title:        req.Title,
		Description:  req.Description,
		ImageURL:     req.ImageURL,
		RankRequired: req.RankRequired,
	}

	err = h.service.AddPrize(c.Request.Context(), userID, challengeID, prize)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Challenge not found")
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "Not authorized to add prizes to this challenge")
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, prize.ToResponse())
}

func (h *ChallengeHandler) UpdatePrize(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	challengeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	prizeID, err := uuid.Parse(c.Param("prize_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_PRIZE_ID", "Invalid prize ID")
		return
	}

	var req prizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	prize := &models.Prize{
		Title:        req.Title,
		Description:  req.Description,
		ImageURL:     req.ImageURL,
		RankRequired: req.RankRequired,
	}

	err = h.service.UpdatePrize(c.Request.Context(), userID, challengeID, prizeID, prize)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Prize or challenge not found")
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "Not authorized to update prizes for this challenge")
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Prize updated successfully"})
}

func (h *ChallengeHandler) DeletePrize(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	challengeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	prizeID, err := uuid.Parse(c.Param("prize_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_PRIZE_ID", "Invalid prize ID")
		return
	}

	err = h.service.DeletePrize(c.Request.Context(), userID, challengeID, prizeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Prize or challenge not found")
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "Not authorized to delete prizes for this challenge")
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Prize deleted successfully"})
}

func (h *ChallengeHandler) GetPrizes(c *gin.Context) {
	challengeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	prizes, err := h.service.GetPrizesByChallengeID(c.Request.Context(), challengeID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch prizes")
		return
	}

	responses := make([]models.PrizeResponse, len(prizes))
	for i, p := range prizes {
		responses[i] = p.ToResponse()
	}

	response.Success(c, http.StatusOK, responses)
}

func (h *ChallengeHandler) PublishChallenge(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	idParam := c.Param("id")
	challengeID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	err = h.service.PublishChallenge(c.Request.Context(), userID, challengeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Challenge not found")
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "Not authorized to publish this challenge")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Challenge published successfully"})
}

func (h *ChallengeHandler) UploadCoverImage(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	idParam := c.Param("id")
	challengeID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "MISSING_FILE", "Image file is required")
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		response.Error(c, http.StatusBadRequest, "INVALID_TYPE", "Only JPG and PNG are allowed")
		return
	}

	f, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "FILE_ERROR", "Failed to open file")
		return
	}
	defer f.Close()

	contentType := "image/jpeg"
	if ext == ".png" {
		contentType = "image/png"
	}

	imageURL, err := h.service.UploadCoverImage(c.Request.Context(), userID, challengeID, f, contentType)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Challenge not found")
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "Not authorized to upload image for this challenge")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to upload cover image")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"image_url": imageURL,
	})
}

func (h *ChallengeHandler) GetChallenge(c *gin.Context) {
	idParam := c.Param("id")
	challengeID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	challenge, err := h.service.GetChallenge(c.Request.Context(), challengeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Challenge not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch challenge")
		return
	}

	response.Success(c, http.StatusOK, challenge.ToResponse())
}

func (h *ChallengeHandler) ListChallenges(c *gin.Context) {
	statusStr := c.Query("status")
	var status *models.ChallengeStatus
	if statusStr != "" {
		s := models.ChallengeStatus(statusStr)
		status = &s
	}

	challenges, err := h.service.ListChallenges(c.Request.Context(), status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list challenges")
		return
	}

	responses := make([]models.ChallengeResponse, len(challenges))
	for i, c := range challenges {
		responses[i] = c.ToResponse()
	}

	response.Success(c, http.StatusOK, responses)
}

func (h *ChallengeHandler) JoinChallenge(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	idParam := c.Param("id")
	challengeID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	err = h.service.JoinChallenge(c.Request.Context(), userID, challengeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Challenge not found")
			return
		}
		if errors.Is(err, domain.ErrAlreadyExists) {
			response.Error(c, http.StatusConflict, "ALREADY_JOINED", "You are already a participant of this challenge")
			return
		}
		if errors.Is(err, domain.ErrChallengeFull) {
			response.Error(c, http.StatusGone, "FULL", "Challenge is already full")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to join challenge")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Successfully joined challenge"})
}

func (h *ChallengeHandler) LeaveChallenge(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	idParam := c.Param("id")
	challengeID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	err = h.service.LeaveChallenge(c.Request.Context(), userID, challengeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Participation not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Successfully left challenge"})
}

func (h *ChallengeHandler) SubmitDailyLog(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	challengeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	var req submitDailyLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	log, err := h.logService.SubmitDailyLog(c.Request.Context(), userID, challengeID, services.SubmitDailyLogInput{
		LogDate:           req.LogDate,
		Steps:             req.Steps,
		Calories:          req.Calories,
		ActiveMinutes:     req.ActiveMinutes,
		HealthKitDataHash: req.HealthKitDataHash,
		SourceBundleIDs:   req.SourceBundleIDs,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Challenge or participation not found")
		case errors.Is(err, domain.ErrAlreadyExists):
			response.Error(c, http.StatusConflict, "ALREADY_LOGGED", "Daily log already submitted for this day")
		case errors.Is(err, domain.ErrInvalidInput), errors.Is(err, domain.ErrBadRequest):
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to submit daily log")
		}
		return
	}

	response.Success(c, http.StatusCreated, log.ToResponse())
}

func (h *ChallengeHandler) GetDailyLogs(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	challengeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	logs, err := h.logService.GetDailyLogsWithAggregation(c.Request.Context(), userID, challengeID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Challenge or participation not found")
		case errors.Is(err, domain.ErrInvalidInput):
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch daily logs")
		}
		return
	}

	response.Success(c, http.StatusOK, logs)
}

func (h *ChallengeHandler) GetMyProgress(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	challengeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	progress, err := h.logService.GetMyProgress(c.Request.Context(), userID, challengeID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Challenge or participation not found")
		case errors.Is(err, domain.ErrInvalidInput):
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch my progress")
		}
		return
	}

	response.Success(c, http.StatusOK, progress)
}

func (h *ChallengeHandler) GetLeaderboard(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	challengeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	leaderboardType := c.DefaultQuery("type", "absolute")
	if leaderboardType != "absolute" {
		response.Error(c, http.StatusBadRequest, "INVALID_TYPE", "Only type=absolute is supported on this endpoint")
		return
	}

	leaderboard, err := h.leaderboardService.GetAbsoluteLeaderboard(c.Request.Context(), challengeID, userID, 20)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Challenge or leaderboard entry not found")
		case errors.Is(err, domain.ErrInvalidInput):
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch leaderboard")
		}
		return
	}

	response.Success(c, http.StatusOK, leaderboard)
}

func (h *ChallengeHandler) GetRelativeLeaderboard(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	challengeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID")
		return
	}

	leaderboard, err := h.leaderboardService.GetRelativeLeaderboard(c.Request.Context(), challengeID, userID, 2)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Challenge or leaderboard entry not found")
		case errors.Is(err, domain.ErrInvalidInput):
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch relative leaderboard")
		}
		return
	}

	response.Success(c, http.StatusOK, leaderboard)
}
