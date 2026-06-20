package repository

import (
	"context"

	"teamtask/internal/domain"
)

type TaskCommentRepository interface {
	Create(ctx context.Context, c *domain.TaskComment) (int64, error)
	ListByTaskID(ctx context.Context, taskID int64) ([]domain.TaskComment, error)
}
