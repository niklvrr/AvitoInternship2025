package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
	"go.uber.org/zap"
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
	Create(ctx context.Context, dto *dto.CreatPrDTO, prReviewers []string) (*result.PrResult, error)
	Merge(ctx context.Context, dto *dto.MergePrDTO) (*result.PrResult, error)
	Reassign(ctx context.Context, dto *dto.ReassignPrDTO) (*result.ReassignResult, error)
	SelectPotentialReviewers(ctx context.Context, userId string) ([]*domain.User, error)
}

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
	authorId, err := normalizeID(req.AuthorId, "author_id")
	if err != nil {
		return nil, err
	}
	s.log.Info("create PR request accepted",
		zap.String("pr_id", req.PrId),
		zap.String("author_id", authorId),
	)

	// Читаем всех членов команды автора
	potentialReviewers, err := s.repo.SelectPotentialReviewers(ctx, authorId)
	if err != nil {
		s.log.Error("failed to load potential reviewers",
			zap.String("author_id", authorId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%w: %w", createError, err)
	}

	// Ищем до двух активных ревьюеров, исключая автора
	reviewers, err := findReviewers(potentialReviewers, authorId, reviewerCountForCreate)
	if err != nil {
		s.log.Warn("no reviewers available",
			zap.String("author_id", authorId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%w: %w", createError, err)
	}

	prId, err := normalizeID(req.PrId, "pull_request_id")
	if err != nil {
		return nil, err
	}

	dto := &dto.CreatPrDTO{
		PrId:     prId,
		PrName:   req.PrName,
		AuthorId: authorId,
	}

	res, err := s.repo.Create(ctx, dto, reviewers)
	if err != nil {
		s.log.Error("failed to create PR",
			zap.String("pr_id", prId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%w: %w", createError, err)
	}

	// Готовим список назначенных ревьюеров для ответа
	var assignedReviewers []string
	for _, reviewer := range res.AssignedReviewers {
		assignedReviewers = append(assignedReviewers, reviewer)
	}

	s.log.Info("PR created",
		zap.String("pr_id", res.Id),
		zap.Strings("assigned_reviewers", assignedReviewers),
	)

	return &response.CreateResponse{
		PrId:              res.Id,
		PrName:            res.Name,
		AuthorId:          res.AuthorId,
		Status:            res.Status,
		AssignedReviewers: assignedReviewers,
		CreatedAt:         formatTime(res.CreatedAt),
		MergedAt:          formatTimePtr(res.MergedAt),
	}, nil
}

func (s *PrService) Merge(ctx context.Context, req *request.MergeRequest) (*response.MergeResponse, error) {
	prId, err := normalizeID(req.PrId, "pull_request_id")
	if err != nil {
		return nil, err
	}
	s.log.Info("merge PR request accepted", zap.String("pr_id", prId))

	dto := &dto.MergePrDTO{
		PrId: prId,
	}

	res, err := s.repo.Merge(ctx, dto)
	if err != nil {
		s.log.Error("failed to merge PR",
			zap.String("pr_id", prId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%w: %w", mergeError, err)
	}

	// Готовим список ревьюеров
	var assignedReviewers []string
	for _, reviewer := range res.AssignedReviewers {
		assignedReviewers = append(assignedReviewers, reviewer)
	}

	s.log.Info("PR merged",
		zap.String("pr_id", res.Id),
		zap.String("status", res.Status),
	)

	return &response.MergeResponse{
		PrId:              res.Id,
		PrName:            res.Name,
		AuthorId:          res.AuthorId,
		Status:            res.Status,
		AssignedReviewers: assignedReviewers,
		CreatedAt:         formatTime(res.CreatedAt),
		MergedAt:          formatTimePtr(res.MergedAt),
	}, nil
}

func (s *PrService) Reassign(ctx context.Context, req *request.ReassignRequest) (*response.ReassignResponse, error) {
	prId, err := normalizeID(req.PrId, "pull_request_id")
	if err != nil {
		return nil, err
	}

	// Парсим идентификатор старого ревьюера
	oldReviewerId, err := normalizeID(req.OldUserId, "old_user_id")
	if err != nil {
		return nil, err
	}
	s.log.Info("reassign reviewer request accepted",
		zap.String("pr_id", prId),
		zap.String("old_user_id", oldReviewerId),
	)

	// Читаем всех членов команды старого ревьюера
	potentialReviewers, err := s.repo.SelectPotentialReviewers(ctx, oldReviewerId)
	if err != nil {
		s.log.Error("failed to load team members for reassign",
			zap.String("old_user_id", oldReviewerId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%w: %w", reassignError, err)
	}

	// Ищем нового активного ревьюера, исключая старого
	newReviewer, err := findReviewers(potentialReviewers, oldReviewerId, reviewerCountForReassign)
	if err != nil {
		s.log.Warn("no replacement reviewer available",
			zap.String("pr_id", prId),
			zap.String("old_user_id", oldReviewerId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%w: %w", reassignError, err)
	}
	newReviewerId := newReviewer[0]

	// Собираем dto для репозитория
	dto := &dto.ReassignPrDTO{
		PrId:          prId,
		OldReviewerId: oldReviewerId,
		ReplacedBy:    newReviewerId,
	}

	res, err := s.repo.Reassign(ctx, dto)
	if err != nil {
		s.log.Error("failed to reassign reviewer",
			zap.String("pr_id", prId),
			zap.String("old_user_id", oldReviewerId),
			zap.String("new_reviewer_id", newReviewerId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%w: %w", reassignError, err)
	}

	// Готовим список ревьюеров для ответа
	var assignedReviewers []string
	for _, reviewer := range res.Pr.AssignedReviewers {
		assignedReviewers = append(assignedReviewers, reviewer)
	}

	s.log.Info("reviewer reassigned",
		zap.String("pr_id", res.Pr.Id),
		zap.Strings("assigned_reviewers", assignedReviewers),
		zap.String("replaced_by", res.ReplacedBy),
	)

	return &response.ReassignResponse{
		PrId:              res.Pr.Id,
		PrName:            res.Pr.Name,
		AuthorId:          res.Pr.AuthorId,
		Status:            res.Pr.Status,
		AssignedReviewers: assignedReviewers,
		ReplacedBy:        res.ReplacedBy,
		CreatedAt:         formatTime(res.Pr.CreatedAt),
		MergedAt:          formatTimePtr(res.Pr.MergedAt),
	}, nil
}

func findReviewers(potentialReviewers []*domain.User, excludedId string, reviewerCount int) ([]string, error) {
	var reviewers []*domain.User
	for _, potentialReviewer := range potentialReviewers {
		if potentialReviewer == nil {
			continue
		}

		if potentialReviewer.Id == excludedId {
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

	result := make([]string, 0, reviewerCount)
	for i := 0; i < reviewerCount; i++ {
		result = append(result, reviewers[i].Id)
	}

	return result, nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := formatTime(*t)
	return &formatted
}

func normalizeID(raw, field string) (string, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("%w: %s is empty", incorrectIdError, field)
	}
	return id, nil
}
