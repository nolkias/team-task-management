package mysql

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"teamtask/internal/domain"
)

type TaskRepo struct {
	db *sql.DB
}

func NewTaskRepo(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, t *domain.Task) (int64, error) {
	status := t.Status
	if status == "" {
		status = domain.StatusTodo
	}

	res, err := r.db.ExecContext(ctx,
		`INSERT INTO tasks (team_id, assignee_id, created_by, title, description, status)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		t.TeamID, t.AssigneeID, t.CreatedBy, t.Title, t.Description, status,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *TaskRepo) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	t := &domain.Task{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, team_id, assignee_id, created_by, title, description, status, created_at, updated_at
		 FROM tasks WHERE id = ?`,
		id,
	).Scan(&t.ID, &t.TeamID, &t.AssigneeID, &t.CreatedBy, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TaskRepo) UpdateWithHistory(ctx context.Context, t *domain.Task, changes []domain.TaskHistory) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET assignee_id = ?, title = ?, description = ?, status = ? WHERE id = ?`,
		t.AssigneeID, t.Title, t.Description, t.Status, t.ID,
	)
	if err != nil {
		return err
	}

	for _, c := range changes {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO task_history (task_id, changed_by, field_changed, old_value, new_value)
			 VALUES (?, ?, ?, ?, ?)`,
			c.TaskID, c.ChangedBy, c.FieldChanged, c.OldValue, c.NewValue,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *TaskRepo) List(ctx context.Context, f domain.TaskFilter, p domain.PageRequest) ([]domain.Task, int, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "team_id = ?")
	args = append(args, f.TeamID)

	if f.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, f.Status)
	}
	if f.AssigneeID != nil {
		conditions = append(conditions, "assignee_id = ?")
		args = append(args, *f.AssigneeID)
	}

	where := strings.Join(conditions, " AND ")

	var total int
	countQuery := "SELECT COUNT(*) FROM tasks WHERE " + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listArgs := append(append([]interface{}{}, args...), p.PageSize, p.Offset())
	listQuery := `SELECT id, team_id, assignee_id, created_by, title, description, status, created_at, updated_at
		 FROM tasks WHERE ` + where + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.TeamID, &t.AssigneeID, &t.CreatedBy, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *TaskRepo) TopCreatorsPerTeamThisMonth(ctx context.Context) ([]domain.TopCreatorRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT team_id, user_id, user_name, tasks_created, rnk
		 FROM (
			SELECT
				tk.team_id AS team_id,
				tk.created_by AS user_id,
				u.name AS user_name,
				COUNT(*) AS tasks_created,
				RANK() OVER (PARTITION BY tk.team_id ORDER BY COUNT(*) DESC) AS rnk
			FROM tasks tk
			JOIN users u ON u.id = tk.created_by
			WHERE tk.created_at >= DATE_FORMAT(NOW(), '%Y-%m-01')
			GROUP BY tk.team_id, tk.created_by, u.name
		 ) ranked
		 WHERE rnk <= 3
		 ORDER BY team_id, rnk`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.TopCreatorRow
	for rows.Next() {
		var row domain.TopCreatorRow
		if err := rows.Scan(&row.TeamID, &row.UserID, &row.UserName, &row.TasksCreated, &row.Rank); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *TaskRepo) OrphanedAssignees(ctx context.Context) ([]domain.OrphanedAssigneeRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT tk.id, tk.title, tk.team_id, tk.assignee_id
		 FROM tasks tk
		 LEFT JOIN team_members tm ON tm.team_id = tk.team_id AND tm.user_id = tk.assignee_id
		 WHERE tk.assignee_id IS NOT NULL AND tm.user_id IS NULL`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.OrphanedAssigneeRow
	for rows.Next() {
		var row domain.OrphanedAssigneeRow
		if err := rows.Scan(&row.TaskID, &row.TaskTitle, &row.TeamID, &row.AssigneeID); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
