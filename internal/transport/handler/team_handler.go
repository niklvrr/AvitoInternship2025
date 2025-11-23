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

type TeamService interface {
	Add(ctx context.Context, req *request.AddTeamRequest) (*response.AddTeamResponse, error)
	Get(ctx context.Context, req *request.GetTeamRequest) (*response.GetTeamResponse, error)
}

type TeamHandler struct {
	svc TeamService
	log *zap.Logger
}

func NewTeamHandler(svc TeamService, log *zap.Logger) *TeamHandler {
	return &TeamHandler{
		svc: svc,
		log: log,
	}
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	h.log.Info("addTeam request received",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	// Парсим json в модель AddTeamRequest
	var req request.AddTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request body", zap.Error(err))
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Валидация
	if req.TeamName == "" {
		h.log.Warn("validation failed: team_name is empty")
		statusCode, errResp := transport.HandleError(usecase.ErrInvalidInput)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Вызов сервиса
	resp, err := h.svc.Add(r.Context(), &req)
	if err != nil {
		h.log.Error("failed to add team",
			zap.String("team_name", req.TeamName),
			zap.Int("members_count", len(req.Members)),
			zap.Error(err),
		)
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	h.log.Info("team added successfully",
		zap.String("team_name", resp.TeamName),
		zap.Int("members_count", len(resp.Members)),
	)

	// Формируем ответ по формату openapi
	members := make([]map[string]interface{}, 0, len(resp.Members))
	for _, member := range resp.Members {
		members = append(members, map[string]interface{}{
			"user_id":   member.Id,
			"username":  member.Name,
			"is_active": member.IsActive,
		})
	}
	response := map[string]interface{}{
		"team": map[string]interface{}{
			"team_name": resp.TeamName,
			"members":   members,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	h.log.Info("getTeam request received",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	// Получаем team_name из query параметров
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.log.Warn("validation failed: team_name query parameter is empty")
		statusCode, errResp := transport.HandleError(usecase.ErrInvalidInput)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Формируем запрос
	req := request.GetTeamRequest{
		TeamName: teamName,
	}

	// Вызываем сервис
	resp, err := h.svc.Get(r.Context(), &req)
	if err != nil {
		h.log.Error("failed to get team",
			zap.String("team_name", teamName),
			zap.Error(err),
		)
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	h.log.Info("team retrieved successfully",
		zap.String("team_name", resp.TeamName),
		zap.Int("members_count", len(resp.Members)),
	)

	// Формируем ответ по формату openapi
	members := make([]map[string]interface{}, 0, len(resp.Members))
	for _, member := range resp.Members {
		members = append(members, map[string]interface{}{
			"user_id":   member.Id,
			"username":  member.Name,
			"is_active": member.IsActive,
		})
	}
	response := map[string]interface{}{
		"team_name": resp.TeamName,
		"members":   members,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
