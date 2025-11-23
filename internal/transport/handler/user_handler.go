package handler

import (
	"context"
	"encoding/json"
	"github.com/niklvrr/AvitoInternship2025/internal/usecase/service"
	"net/http"

	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
	"go.uber.org/zap"
)

type UserService interface {
	SetIsActive(ctx context.Context, req *request.SetIsActiveRequest) (*response.SetIsActiveResponse, error)
	GetReview(ctx context.Context, req *request.GetReviewRequest) (*response.GetReviewResponse, error)
}

type UserHandler struct {
	svc UserService
	log *zap.Logger
}

func NewUserHandler(svc UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{
		svc: svc,
		log: log,
	}
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	h.log.Info("setIsActive request received",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	// Парсим json в модель SetIsActiveRequest
	var req request.SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request body", zap.Error(err))
		statusCode, errResp := HandleError(err)
		WriteError(w, statusCode, errResp)
		return
	}

	// Валидация
	if req.UserId == "" {
		h.log.Warn("validation failed: user_id is empty")
		statusCode, errResp := HandleError(service.ErrInvalidInput)
		WriteError(w, statusCode, errResp)
		return
	}

	// Вызов сервиса
	resp, err := h.svc.SetIsActive(r.Context(), &req)
	if err != nil {
		h.log.Error("failed to set user active status",
			zap.String("user_id", req.UserId),
			zap.Bool("is_active", req.IsActive),
			zap.Error(err),
		)
		statusCode, errResp := HandleError(err)
		WriteError(w, statusCode, errResp)
		return
	}

	h.log.Info("user active status updated",
		zap.String("user_id", resp.UserId),
		zap.Bool("is_active", resp.IsActive),
	)

	// Формируем ответ по openapi
	response := map[string]interface{}{
		"user": resp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	h.log.Info("getReview request received",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	// Получаем user_id из query параметров
	userId := r.URL.Query().Get("user_id")
	if userId == "" {
		h.log.Warn("validation failed: user_id query parameter is empty")
		statusCode, errResp := HandleError(service.ErrInvalidInput)
		WriteError(w, statusCode, errResp)
		return
	}

	// Формируем запрос
	req := request.GetReviewRequest{
		UserId: userId,
	}

	// Вызываем сервис
	resp, err := h.svc.GetReview(r.Context(), &req)
	if err != nil {
		h.log.Error("failed to get user reviews",
			zap.String("user_id", userId),
			zap.Error(err),
		)
		statusCode, errResp := HandleError(err)
		WriteError(w, statusCode, errResp)
		return
	}

	h.log.Info("user reviews retrieved",
		zap.String("user_id", resp.UserId),
		zap.Int("pull_requests_count", len(resp.Prs)),
	)

	// Формируем ответ по openapi
	pullRequests := make([]map[string]interface{}, 0, len(resp.Prs))
	for _, pr := range resp.Prs {
		pullRequests = append(pullRequests, map[string]interface{}{
			"pull_request_id":   pr.Id,
			"pull_request_name": pr.Name,
			"author_id":         pr.AuthorId,
			"status":            pr.Status,
		})
	}
	response := map[string]interface{}{
		"user_id":       resp.UserId,
		"pull_requests": pullRequests,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
