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

type PrService interface {
	Create(ctx context.Context, req *request.CreateRequest) (*response.CreateResponse, error)
	Merge(ctx context.Context, req *request.MergeRequest) (*response.MergeResponse, error)
	Reassign(ctx context.Context, req *request.ReassignRequest) (*response.ReassignResponse, error)
}

type PrHandler struct {
	svc PrService
	log *zap.Logger
}

func NewPrHandler(svc PrService, log *zap.Logger) *PrHandler {
	return &PrHandler{
		svc: svc,
		log: log,
	}
}

func (h *PrHandler) CreatePr(w http.ResponseWriter, r *http.Request) {
	// Парсим json в модель CreateRequest
	var req request.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Валидация
	if req.PrId == "" || req.PrName == "" || req.AuthorId == "" {
		statusCode, errResp := transport.HandleError(usecase.ErrInvalidInput)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Вызов сервиса
	resp, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Формируем запрос по openapi
	response := map[string]interface{}{
		"pr": resp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *PrHandler) MergePr(w http.ResponseWriter, r *http.Request) {
	// Парсим json в модель MergeRequest
	var req request.MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Валидация
	if req.PrId == "" {
		statusCode, errResp := transport.HandleError(usecase.ErrInvalidInput)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Вызов сервиса
	resp, err := h.svc.Merge(r.Context(), &req)
	if err != nil {
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Формируем ответ по формату openapi
	response := map[string]interface{}{
		"pr": resp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *PrHandler) ReassignPr(w http.ResponseWriter, r *http.Request) {
	// Парсим json в модель ReassignRequest
	var req request.ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Валидация
	if req.PrId == "" || req.OldUserId == "" {
		statusCode, errResp := transport.HandleError(usecase.ErrInvalidInput)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Вызов сервиса
	resp, err := h.svc.Reassign(r.Context(), &req)
	if err != nil {
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Формируем ответ по формату openapi
	response := map[string]interface{}{
		"pr":          resp,
		"replaced_by": resp.ReplacedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
