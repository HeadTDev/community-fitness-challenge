package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

type JSONResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      Meta        `json:"meta"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
}

func Success(c *gin.Context, status int, data interface{}) {
	rid, _ := c.Get("X-Request-ID")
	requestID, _ := rid.(string) // Safe assertion, defaults to empty string

	c.JSON(status, JSONResponse{
		Success: true,
		Data:    data,
		Meta: Meta{
			Timestamp: time.Now(),
			RequestID: requestID,
		},
	})
}

func Error(c *gin.Context, status int, code, message string) {
	rid, _ := c.Get("X-Request-ID")
	requestID, _ := rid.(string) // Safe assertion, defaults to empty string

	c.AbortWithStatusJSON(status, JSONResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
		Meta: Meta{
			Timestamp: time.Now(),
			RequestID: requestID,
		},
	})
}
