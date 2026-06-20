package repository

import (
	"context"

	"teamtask/internal/domain"
)

type TaskRepository interface {
	Create(ctx context.Context, t *domain.Task) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Task, error)
	UpdateWithHistory(ctx context.Context, t *domain.Task, changes []domain.TaskHistory) error
	List(ctx context.Context, f domain.TaskFilter, p domain.PageRequest) ([]domain.Task, int, error)
	TopCreatorsPerTeamThisMonth(ctx context.Context) ([]domain.TopCreatorRow, error)
	OrphanedAssignees(ctx context.Context) ([]domain.OrphanedAssigneeRow, error)
}
