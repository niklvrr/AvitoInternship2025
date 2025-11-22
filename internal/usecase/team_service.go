package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
)

var (
	addTeamError = errors.New("add team error")
	getTeamError = errors.New("get team error")
)

// Интерфейс репозитория
type TeamRepository interface {
	Add(ctx context.Context, dto *dto.AddTeamDTO) (*result.AddTeamResult, error)
	Get(ctx context.Context, dto *dto.GetTeamDTO) (*result.GetTeamResult, error)
}

type TeamService struct {
	repo TeamRepository
}

func NewTeamService(repo TeamRepository) *TeamService {
	return &TeamService{repo: repo}
}

func (s *TeamService) Add(ctx context.Context, req *request.AddTeamRequest) (*response.AddTeamResponse, error) {
	// Собираем dto
	dto := &dto.AddTeamDTO{
		TeamName: req.TeamName,
		Members:  req.Members,
	}

	// Запрос в бд
	res, err := s.repo.Add(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", addTeamError, err)
	}

	// Ответ
	return &response.AddTeamResponse{
		TeamName: res.TeamName,
		Members:  res.Members,
	}, nil
}

func (s *TeamService) Get(ctx context.Context, req *request.GetTeamRequest) (*response.GetTeamResponse, error) {
	// Собираем dto
	dto := &dto.GetTeamDTO{
		TeamName: req.TeamName,
	}

	// Запрос в бд
	res, err := s.repo.Get(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", getTeamError, err)
	}

	// Ответ
	return &response.GetTeamResponse{
		TeamName: res.TeamName,
		Members:  res.Members,
	}, nil
}
