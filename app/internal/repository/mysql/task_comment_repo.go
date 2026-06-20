package mysql

import (
	"context"
	"database/sql"

	"teamtask/internal/domain"
)

type TaskCommentRepo struct {
	db *sql.DB
}

func NewTaskCommentRepo(db *sql.DB) *TaskCommentRepo {
	return &TaskCommentRepo{db: db}
}

func (r *TaskCommentRepo) Create(ctx context.Context, c *domain.TaskComment) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO task_comments (task_id, user_id, body) VALUES (?, ?, ?)`,
		c.TaskID, c.UserID, c.Body,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *TaskCommentRepo) ListByTaskID(ctx context.Context, taskID int64) ([]domain.TaskComment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, task_id, user_id, body, created_at FROM task_comments WHERE task_id = ? ORDER BY created_at ASC`,
		taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.TaskComment
	for rows.Next() {
		var c domain.TaskComment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
