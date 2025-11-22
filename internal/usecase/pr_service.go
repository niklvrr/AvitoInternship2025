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
	SelectAuthorOfPr(ctx context.Context, prId uuid.UUID) (*uuid.UUID, error)
}

// TODO добавить логирование
// TODO сделать pull в develop
// TODO изменить тип id с uuid на string

// Интерфейс репозитория
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
	authorId, err := uuid.Parse(req.AuthorId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", incorrectId, err)
	}

	// Запрос в бд для чтения всех возможных ревьюеров
	potentialReviewers, err := s.repo.SelectPotentialReviewers(ctx, authorId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", createError, err)
	}

	// Находим до 2 ревьюеров с is_active=true, если ревьюеров нет возращаем ошибку
	reviewers, err := findReviewers(potentialReviewers, reviewerCountForCreate, authorId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", createError, err)
	}

	prId, err := uuid.Parse(req.PrId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", incorrectId, err)
	}

	// Собираем dto
	dto := &dto.CreatPrDTO{
		PrId:     prId,
		PrName:   req.PrName,
		AuthorId: authorId,
	}

	// Запрос в бд для создания pull request
	res, err := s.repo.Create(ctx, dto, reviewers)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", createError, err)
	}

	var assignedReviewers []string
	for _, reviewer := range res.AssignedReviewers {
		assignedReviewers = append(assignedReviewers, reviewer.String())
	}

	// Ответ
	return &response.CreateResponse{
		PrId:              res.Id.String(),
		PrName:            res.Name,
		AuthorId:          res.AuthorId.String(),
		Status:            res.Status,
		AssignedReviewers: assignedReviewers,
	}, nil
}

func (s *PrService) Merge(ctx context.Context, req *request.MergeRequest) (*response.MergeResponse, error) {
	prId, err := uuid.Parse(req.PrId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", incorrectId, err)
	}

	// Собираем dto
	dto := &dto.MergePrDTO{
		PrId: prId,
	}

	// Запрос в бд для изменения статуса pr
	res, err := s.repo.Merge(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", mergeError, err)
	}

	var assignedReviewers []string
	for _, reviewer := range res.AssignedReviewers {
		assignedReviewers = append(assignedReviewers, reviewer.String())
	}

	// Ответ
	return &response.MergeResponse{
		PrId:              res.Id.String(),
		PrName:            res.Name,
		AuthorId:          res.AuthorId.String(),
		Status:            res.Status,
		AssignedReviewers: assignedReviewers,
	}, nil
}

func (s *PrService) Reassign(ctx context.Context, req *request.ReassignRequest) (*response.ReassignResponse, error) {
	prId, err := uuid.Parse(req.PrId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", incorrectId, err)
	}

	// Запрос в бд для чтения всех возможных ревьюеров
	potentialReviewers, err := s.repo.SelectPotentialReviewers(ctx, prId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", reassignError, err)
	}

	// Запрос в бд для чтения author_id для pr
	authorId, err := s.repo.SelectAuthorOfPr(ctx, prId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", reassignError, err)
	}

	// Находим 1 ревьюера с is_active=true, если ревьюеров нет возращаем ошибку
	newReviewer, err := findReviewers(potentialReviewers, reviewerCountForReassign, *authorId, prId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", reassignError, err)
	}
	newReviewerId := *newReviewer[0]

	oldReviewerId, err := uuid.Parse(req.OldReviewerId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", incorrectId, err)
	}

	// Собираем dto
	dto := &dto.ReassignPrDTO{
		PrId:          prId,
		OldReviewerId: oldReviewerId,
		ReplacedBy:    newReviewerId,
	}

	// Запрос в бд для переназначения
	res, err := s.repo.Reassign(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", reassignError, err)
	}

	var assignedReviewers []string
	for _, reviewer := range res.Pr.AssignedReviewers {
		assignedReviewers = append(assignedReviewers, reviewer.String())
	}

	// Ответ
	return &response.ReassignResponse{
		PrId:              res.Pr.Id.String(),
		PrName:            res.Pr.Name,
		AuthorId:          res.Pr.AuthorId.String(),
		Status:            res.Pr.Status,
		AssignedReviewers: assignedReviewers,
		ReplacedBy:        res.ReplacedBy.String(),
	}, nil
}

// вспомогательная функция поиска подходящих ревьюеров
func findReviewers(potentialReviewers []*domain.User, reviewerCount int, ids ...uuid.UUID) ([]*uuid.UUID, error) {
	var reviewers []*domain.User
	for _, potentialReviewer := range potentialReviewers {
		if potentialReviewer == nil {
			continue
		}

		// Фильтруем ревьеров по id (например на случай, если reviewer_id=author_id)
		flag := false
		for _, id := range ids {
			if id == potentialReviewer.Id {
				flag = true
				break
			}
		}
		if flag {
			continue
		}

		// Фильтуем по is_active
		if potentialReviewer.IsActive == false {
			continue
		}

		// Если ревьюер подходит по всем учловиям то добавляем
		reviewers = append(reviewers, potentialReviewer)
	}

	// Если ревьюеров нет возращаем ошибку
	if len(reviewers) == 0 {
		return nil, noPotentialReviewerError
	}

	// Случайно перемешиваем слайс ревьюеров
	rand.New(rand.NewSource(time.Now().UnixNano())).Shuffle(len(reviewers), func(i, j int) {
		reviewers[i], reviewers[j] = reviewers[j], reviewers[i]
	})

	// меняем reviewerCount, если осталось меньше
	if len(reviewers) < reviewerCount {
		reviewerCount = len(reviewers)
	}

	// отбираем первых reviewerCount ревьюеров
	result := make([]*uuid.UUID, 0, reviewerCount)
	for i := 0; i < reviewerCount; i++ {
		id := reviewers[i].Id
		result = append(result, &id)
	}

	return result, nil
}
