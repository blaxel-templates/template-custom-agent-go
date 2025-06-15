package models

import "time"

// ErrorResponse represents a standard error response format
type ErrorResponse struct {
	Error     string    `json:"error"`
	Code      int       `json:"code"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}
