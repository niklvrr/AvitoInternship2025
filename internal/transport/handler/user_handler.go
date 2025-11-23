package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/niklvrr/AvitoInternship2025/internal/transport"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
	"github.com/niklvrr/AvitoInternship2025/internal/usecase"
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
	// Парсим json в модель SetIsActiveRequest
	var req request.SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Валидация
	if req.UserId == "" {
		statusCode, errResp := transport.HandleError(usecase.ErrInvalidInput)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Вызов сервиса
	resp, err := h.svc.SetIsActive(r.Context(), &req)
	if err != nil {
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Формируем ответ по openapi
	response := map[string]interface{}{
		"user": resp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	// Получаем user_id из query параметров
	userId := r.URL.Query().Get("user_id")
	if userId == "" {
		statusCode, errResp := transport.HandleError(usecase.ErrInvalidInput)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Формируем запрос
	req := request.GetReviewRequest{
		UserId: userId,
	}

	// Вызываем сервис
	resp, err := h.svc.GetReview(r.Context(), &req)
	if err != nil {
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

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
