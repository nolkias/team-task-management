package mysql

import (
	"context"
	"database/sql"

	"teamtask/internal/domain"
)

type TaskHistoryRepo struct {
	db *sql.DB
}

func NewTaskHistoryRepo(db *sql.DB) *TaskHistoryRepo {
	return &TaskHistoryRepo{db: db}
}

func (r *TaskHistoryRepo) ListByTaskID(ctx context.Context, taskID int64) ([]domain.TaskHistory, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, task_id, changed_by, field_changed, old_value, new_value, changed_at
		 FROM task_history WHERE task_id = ? ORDER BY changed_at ASC`,
		taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.TaskHistory
	for rows.Next() {
		var h domain.TaskHistory
		if err := rows.Scan(&h.ID, &h.TaskID, &h.ChangedBy, &h.FieldChanged, &h.OldValue, &h.NewValue, &h.ChangedAt); err != nil {
			return nil, err
		}
		result = append(result, h)
	}
	return result, rows.Err()
}
