package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

var (
	createError              = errors.New("create pull request error")
	mergeError               = errors.New("merge pull request error")
	reassignError            = errors.New("reassigning pull request reviewer error")
	noPotentialReviewerError = errors.New("no active reviewer available")
)

const (
	reviewerCountForCreate   = 2
	reviewerCountForReassign = 1
)

// Интерфейс репозитория
type PrRepository interface {
	Create(ctx context.Context, dto *dto.CreatPrDTO, prReviewers []*uuid.UUID) (*result.PrResult, error)
	Merge(ctx context.Context, dto *dto.MergePrDTO) (*result.PrResult, error)
	Reassign(ctx context.Context, dto *dto.ReassignPrDTO) (*result.ReassignResult, error)
	SelectPotentialReviewers(ctx context.Context, userId uuid.UUID) ([]*domain.User, error)
}

// TODO добавить логирование
// TODO добавить комментарии
// TODO сделать pull в develop

type PrService struct {
	repo PrRepository
	log  *zap.Logger
}

func NewPrService(repo PrRepository, log *zap.Logger) *PrService {
	return &PrService{
		repo: repo,
		log:  log,
	}
}

func (s *PrService) Create(ctx context.Context, req *request.CreateRequest) (*response.CreateResponse, error) {
	// Парсим автора
	authorId, err := uuid.Parse(req.AuthorId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", incorrectIdError, err)
	}

	// Берем всех участников команды автора
	potentialReviewers, err := s.repo.SelectPotentialReviewers(ctx, authorId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", createError, err)
	}

	// Ищем активных ревьюеров и исключаем автора
	reviewers, err := findReviewers(potentialReviewers, reviewerCountForCreate, authorId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", createError, err)
	}

	// Парсим идентификатор PR
	prId, err := uuid.Parse(req.PrId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", incorrectIdError, err)
	}

	// Формируем DTO для репозитория
	dto := &dto.CreatPrDTO{
		PrId:     prId,
		PrName:   req.PrName,
		AuthorId: authorId,
	}

	// Создаем PR в базе
	res, err := s.repo.Create(ctx, dto, reviewers)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", createError, err)
	}

	// Собираем ответ
	var assignedReviewers []string
	for _, reviewer := range res.AssignedReviewers {
		assignedReviewers = append(assignedReviewers, reviewer.String())
	}

	return &response.CreateResponse{
		PrId:              res.Id.String(),
		PrName:            res.Name,
		AuthorId:          res.AuthorId.String(),
		Status:            res.Status,
		AssignedReviewers: assignedReviewers,
		CreatedAt:         formatTimeValue(res.CreatedAt),
		MergedAt:          formatTimePtr(res.MergedAt),
	}, nil
}

func (s *PrService) Merge(ctx context.Context, req *request.MergeRequest) (*response.MergeResponse, error) {
	// Парсим идентификатор PR
	prId, err := uuid.Parse(req.PrId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", incorrectIdError, err)
	}

	// Формируем DTO
	dto := &dto.MergePrDTO{
		PrId: prId,
	}

	// Обновляем статус PR
	res, err := s.repo.Merge(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", mergeError, err)
	}

	// Собираем список ревьюеров
	var assignedReviewers []string
	for _, reviewer := range res.AssignedReviewers {
		assignedReviewers = append(assignedReviewers, reviewer.String())
	}

	// Готовим ответ
	return &response.MergeResponse{
		PrId:              res.Id.String(),
		PrName:            res.Name,
		AuthorId:          res.AuthorId.String(),
		Status:            res.Status,
		AssignedReviewers: assignedReviewers,
		MergedAt:          res.MergedAt.String(),
	}, nil
}

func (s *PrService) Reassign(ctx context.Context, req *request.ReassignRequest) (*response.ReassignResponse, error) {
	// Парсим идентификатор PR
	prId, err := uuid.Parse(req.PrId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", incorrectIdError, err)
	}

	// Парсим идентификатор заменяемого ревьюера
	oldReviewerId, err := uuid.Parse(req.OldReviewerId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", incorrectIdError, err)
	}

	// Берем всех участников команды заменяемого ревьюера
	potentialReviewers, err := s.repo.SelectPotentialReviewers(ctx, oldReviewerId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", reassignError, err)
	}

	// Ищем новую кандидатуру
	newReviewer, err := findReviewers(potentialReviewers, reviewerCountForReassign, oldReviewerId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", reassignError, err)
	}
	newReviewerId := *newReviewer[0]

	// Формируем DTO
	dto := &dto.ReassignPrDTO{
		PrId:          prId,
		OldReviewerId: oldReviewerId,
		NewReviewerId: newReviewerId,
	}

	// Проводим замену в базе
	res, err := s.repo.Reassign(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", reassignError, err)
	}

	// Собираем ответ
	var assignedReviewers []string
	for _, reviewer := range res.Pr.AssignedReviewers {
		assignedReviewers = append(assignedReviewers, reviewer.String())
	}

	return &response.ReassignResponse{
		PrId:              res.Pr.Id.String(),
		PrName:            res.Pr.Name,
		AuthorId:          res.Pr.AuthorId.String(),
		Status:            res.Pr.Status,
		AssignedReviewers: assignedReviewers,
		ReplacedBy:        res.ReplacedBy.String(),
		CreatedAt:         formatTimeValue(res.Pr.CreatedAt),
		MergedAt:          formatTimePtr(res.Pr.MergedAt),
	}, nil
}

func findReviewers(potentialReviewers []*domain.User, authorId uuid.UUID, reviewerCount int) ([]*uuid.UUID, error) {
	var reviewers []*domain.User
	for _, potentialReviewer := range potentialReviewers {
		if potentialReviewer == nil {
			continue
		}

		if potentialReviewer.Id == authorId {
			continue
		}

		if potentialReviewer.IsActive == false {
			continue
		}

		reviewers = append(reviewers, potentialReviewer)
	}

	if len(reviewers) == 0 {
		return nil, noPotentialReviewerError
	}

	rand.New(rand.NewSource(time.Now().UnixNano())).Shuffle(len(reviewers), func(i, j int) {
		reviewers[i], reviewers[j] = reviewers[j], reviewers[i]
	})

	if len(reviewers) < reviewerCount {
		reviewerCount = len(reviewers)
	}

	result := make([]*uuid.UUID, 0, reviewerCount)
	for i := 0; i < reviewerCount; i++ {
		id := reviewers[i].Id
		result = append(result, &id)
	}

	return result, nil
}
