package handler

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type HealthHandler struct {
	log *zap.Logger
}

func NewHealthHandler(log *zap.Logger) *HealthHandler {
	return &HealthHandler{
		log: log,
	}
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("health check requested",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	response := map[string]string{
		"status": "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
