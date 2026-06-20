package apperror

import (
	"errors"
	"net/http"

	"teamtask/internal/domain"
)

func StatusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrTeamNotFound),
		errors.Is(err, domain.ErrTaskNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrUserAlreadyExists),
		errors.Is(err, domain.ErrAlreadyTeamMember):
		return http.StatusConflict
	case errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrNotTeamMember),
		errors.Is(err, domain.ErrInsufficientRole):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrAssigneeNotMember),
		errors.Is(err, domain.ErrInvalidStatus):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
