package handler

import (
	"encoding/json"
	"errors"
	"github.com/niklvrr/AvitoInternship2025/internal/usecase/service"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HandleError маппит доменные ошибки на HTTP коды и ErrorResponse
func HandleError(err error) (int, ErrorResponse) {
	if err == nil {
		return http.StatusOK, ErrorResponse{}
	}

	var domainErr *service.DomainError
	if errors.As(err, &domainErr) {
		// Маппим код ошибки на HTTP статус
		statusCode := mapErrorCodeToHTTPStatus(domainErr.Code)
		return statusCode, ErrorResponse{
			Error: ErrorDetail{
				Code:    domainErr.Code,
				Message: domainErr.Message,
			},
		}
	}

	// Неизвестная ошибка - возвращаем 500
	return http.StatusInternalServerError, ErrorResponse{
		Error: ErrorDetail{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		},
	}
}

// mapErrorCodeToHTTPStatus маппит код ошибки из OpenAPI на HTTP статус
func mapErrorCodeToHTTPStatus(code string) int {
	switch code {
	case "TEAM_EXISTS":
		return http.StatusBadRequest // 400
	case "PR_EXISTS":
		return http.StatusConflict // 409
	case "PR_MERGED":
		return http.StatusConflict // 409
	case "NOT_ASSIGNED":
		return http.StatusConflict // 409
	case "NO_CANDIDATE":
		return http.StatusConflict // 409
	case "NOT_FOUND":
		return http.StatusNotFound // 404
	case "INVALID_INPUT":
		return http.StatusBadRequest // 400
	default:
		return http.StatusInternalServerError // 500
	}
}

// WriteError отправляет ErrorResponse клиенту
func WriteError(w http.ResponseWriter, statusCode int, errResp ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errResp)
}
