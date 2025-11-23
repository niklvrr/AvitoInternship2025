package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
	"go.uber.org/zap"
)

type StatsService interface {
	GetStats(ctx context.Context) (*response.StatsResponse, error)
}

type StatsHandler struct {
	svc StatsService
	log *zap.Logger
}

func NewStatsHandler(svc StatsService, log *zap.Logger) *StatsHandler {
	return &StatsHandler{
		svc: svc,
		log: log,
	}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	h.log.Info("getStats request received",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	resp, err := h.svc.GetStats(r.Context())
	if err != nil {
		h.log.Error("failed to get statistics",
			zap.Error(err),
		)
		statusCode, errResp := HandleError(err)
		WriteError(w, statusCode, errResp)
		return
	}

	h.log.Info("statistics retrieved successfully",
		zap.Int("users_count", len(resp.Users)),
		zap.Int("prs_count", len(resp.PRs)),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
