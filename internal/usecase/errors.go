package usecase

import (
	"errors"
	"fmt"
)

var (
	incorrectIdError = errors.New("incorrect id error")
)

type DomainError struct {
	Code    string
	Message string
	Err     error
}

func WrapError(domainError *DomainError, err error) error {
	return &DomainError{
		Code:    domainError.Code,
		Message: domainError.Message,
		Err:     err,
	}
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %w", e.Message, e.Err)
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

var (
	// NOT_FOUND
	ErrTeamNotFound = &DomainError{
		Code:    "NOT_FOUND",
		Message: "team not found",
	}
	ErrUserNotFound = &DomainError{
		Code:    "NOT_FOUND",
		Message: "user not found",
	}
	ErrPrNotFound = &DomainError{
		Code:    "NOT_FOUND",
		Message: "pull request not found",
	}

	// TEAM_EXISTS
	ErrTeamExists = &DomainError{
		Code:    "TEAM_EXISTS",
		Message: "team_name already exists",
	}

	// PR_EXISTS
	ErrPrExists = &DomainError{
		Code:    "PR_EXISTS",
		Message: "PR id already exists",
	}

	// PR_MERGED
	ErrPrMerged = &DomainError{
		Code:    "PR_MERGED",
		Message: "cannot reassign on merged PR",
	}

	// NOT_ASSIGNED
	ErrReviewerNotAssigned = &DomainError{
		Code:    "NOT_ASSIGNED",
		Message: "reviewer is not assigned to this PR",
	}

	// NO_CANDIDATE
	ErrNoCandidate = &DomainError{
		Code:    "NO_CANDIDATE",
		Message: "no active replacement candidate in team",
	}

	// INVALID_INPUT
	ErrInvalidInput = &DomainError{
		Code:    "INVALID_INPUT",
		Message: "invalid input",
	}
)
