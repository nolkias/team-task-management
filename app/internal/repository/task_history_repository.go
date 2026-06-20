package repository

import (
	"context"

	"teamtask/internal/domain"
)

type TaskHistoryRepository interface {
	ListByTaskID(ctx context.Context, taskID int64) ([]domain.TaskHistory, error)
}
