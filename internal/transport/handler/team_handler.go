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
	// Парсим json в модель AddTeamRequest
	var req request.AddTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Валидация
	if req.TeamName == "" {
		statusCode, errResp := transport.HandleError(usecase.ErrInvalidInput)
		transport.WriteError(w, statusCode, errResp)
		return
	}

	// Вызов сервиса
	resp, err := h.svc.Add(r.Context(), &req)
	if err != nil {
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

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
	// Получаем team_name из query параметров
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
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
		statusCode, errResp := transport.HandleError(err)
		transport.WriteError(w, statusCode, errResp)
		return
	}

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
