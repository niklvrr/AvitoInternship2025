package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/repository"

	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
	"go.uber.org/zap"
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
	log  *zap.Logger
}

func NewTeamService(repo TeamRepository, log *zap.Logger) *TeamService {
	return &TeamService{
		repo: repo,
		log:  log,
	}
}

func (s *TeamService) Add(ctx context.Context, req *request.AddTeamRequest) (*response.AddTeamResponse, error) {
	s.log.Info("add team request accepted", zap.String("team_name", req.TeamName))
	// Собираем dto
	dto := &dto.AddTeamDTO{
		TeamName: req.TeamName,
		Members:  req.Members,
	}

	// Запрос в бд
	res, err := s.repo.Add(ctx, dto)
	if err != nil {
		s.log.Error("failed to add team", zap.String("team_name", req.TeamName), zap.Error(err))

		// Маппим ошибки
		if errors.Is(err, repository.ErrInvalidInput) {
			return nil, WrapError(ErrInvalidInput, err)
		}
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, WrapError(ErrTeamExists, err)
		}
		if errors.Is(err, repository.ErrNotFound) {
			return nil, WrapError(ErrTeamNotFound, err)
		}

		// Неизвестная ошибка
		return nil, fmt.Errorf("%w: %w", addTeamError, err)
	}

	s.log.Info("team added", zap.String("team_name", res.TeamName), zap.Int("members", len(res.Members)))
	// Ответ
	return &response.AddTeamResponse{
		TeamName: res.TeamName,
		Members:  res.Members,
	}, nil
}

func (s *TeamService) Get(ctx context.Context, req *request.GetTeamRequest) (*response.GetTeamResponse, error) {
	s.log.Info("get team request accepted", zap.String("team_name", req.TeamName))
	// Собираем dto
	dto := &dto.GetTeamDTO{
		TeamName: req.TeamName,
	}

	// Запрос в бд
	res, err := s.repo.Get(ctx, dto)
	if err != nil {
		s.log.Error("failed to get team", zap.String("team_name", req.TeamName), zap.Error(err))

		// Маппим ошибки
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, WrapError(ErrTeamExists, err)
		}
		if errors.Is(err, repository.ErrNotFound) {
			return nil, WrapError(ErrTeamNotFound, err)
		}

		// Неизвестная ошибка
		return nil, fmt.Errorf("%w: %w", getTeamError, err)
	}

	s.log.Info("team fetched", zap.String("team_name", res.TeamName), zap.Int("members", len(res.Members)))
	// Ответ
	return &response.GetTeamResponse{
		TeamName: res.TeamName,
		Members:  res.Members,
	}, nil
}
