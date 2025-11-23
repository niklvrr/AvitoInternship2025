package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
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
	h.log.Info("createPr request received",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	// Парсим json в модель CreateRequest
	var req request.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request body", zap.Error(err))
		statusCode, errResp := HandleError(err)
		WriteError(w, statusCode, errResp)
		return
	}

	// Вызов сервиса
	resp, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		h.log.Error("failed to create PR",
			zap.String("pr_id", req.PrId),
			zap.String("author_id", req.AuthorId),
			zap.Error(err),
		)
		statusCode, errResp := HandleError(err)
		WriteError(w, statusCode, errResp)
		return
	}

	h.log.Info("PR created successfully",
		zap.String("pr_id", resp.PrId),
		zap.Strings("assigned_reviewers", resp.AssignedReviewers),
	)

	// Формируем запрос по openapi
	response := map[string]interface{}{
		"pr": resp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *PrHandler) MergePr(w http.ResponseWriter, r *http.Request) {
	h.log.Info("mergePr request received",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	// Парсим json в модель MergeRequest
	var req request.MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request body", zap.Error(err))
		statusCode, errResp := HandleError(err)
		WriteError(w, statusCode, errResp)
		return
	}

	// Вызов сервиса
	resp, err := h.svc.Merge(r.Context(), &req)
	if err != nil {
		h.log.Error("failed to merge PR",
			zap.String("pr_id", req.PrId),
			zap.Error(err),
		)
		statusCode, errResp := HandleError(err)
		WriteError(w, statusCode, errResp)
		return
	}

	h.log.Info("PR merged successfully",
		zap.String("pr_id", resp.PrId),
		zap.String("status", resp.Status),
	)

	// Формируем ответ по формату openapi
	response := map[string]interface{}{
		"pr": resp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *PrHandler) ReassignPr(w http.ResponseWriter, r *http.Request) {
	h.log.Info("reassignPr request received",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	// Парсим json в модель ReassignRequest
	var req request.ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request body", zap.Error(err))
		statusCode, errResp := HandleError(err)
		WriteError(w, statusCode, errResp)
		return
	}

	// Вызов сервиса
	resp, err := h.svc.Reassign(r.Context(), &req)
	if err != nil {
		h.log.Error("failed to reassign PR reviewer",
			zap.String("pr_id", req.PrId),
			zap.String("old_user_id", req.OldUserId),
			zap.Error(err),
		)
		statusCode, errResp := HandleError(err)
		WriteError(w, statusCode, errResp)
		return
	}

	h.log.Info("PR reviewer reassigned successfully",
		zap.String("pr_id", resp.PrId),
		zap.String("replaced_by", resp.ReplacedBy),
	)

	// Формируем ответ по формату openapi
	response := map[string]interface{}{
		"pr":          resp,
		"replaced_by": resp.ReplacedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
