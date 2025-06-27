package models

type APIError struct {
	Error     string `json:"error"`
	ErrorCode string `json:"errorCode"`
}

type ErrorResponse struct {
	Errors []APIError `json:"errors"`
}