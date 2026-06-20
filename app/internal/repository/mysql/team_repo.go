package mysql

import (
	"context"
	"database/sql"
	"errors"

	"teamtask/internal/domain"
)

type TeamRepo struct {
	db *sql.DB
}

func NewTeamRepo(db *sql.DB) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) CreateWithOwner(ctx context.Context, team *domain.Team, ownerID int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO teams (name, created_by) VALUES (?, ?)`,
		team.Name, ownerID,
	)
	if err != nil {
		return 0, err
	}

	teamID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)`,
		teamID, ownerID, domain.RoleOwner,
	)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return teamID, nil
}

func (r *TeamRepo) ListForUser(ctx context.Context, userID int64) ([]domain.Team, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT t.id, t.name, t.created_by, t.created_at
		 FROM teams t
		 JOIN team_members tm ON tm.team_id = t.id
		 WHERE tm.user_id = ?
		 ORDER BY t.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []domain.Team
	for rows.Next() {
		var t domain.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

func (r *TeamRepo) GetMembership(ctx context.Context, teamID, userID int64) (*domain.TeamMember, error) {
	m := &domain.TeamMember{}
	err := r.db.QueryRowContext(ctx,
		`SELECT team_id, user_id, role, joined_at FROM team_members WHERE team_id = ? AND user_id = ?`,
		teamID, userID,
	).Scan(&m.TeamID, &m.UserID, &m.Role, &m.JoinedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotTeamMember
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *TeamRepo) AddMember(ctx context.Context, teamID, userID int64, role domain.Role) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)`,
		teamID, userID, role,
	)
	return err
}

func (r *TeamRepo) CountDistinctMembers(ctx context.Context, teamID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM team_members WHERE team_id = ?`,
		teamID,
	).Scan(&count)
	return count, err
}

func (r *TeamRepo) TeamStats(ctx context.Context) ([]domain.TeamStatsRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			t.id,
			t.name,
			COUNT(DISTINCT tm.user_id) AS member_count,
			COUNT(DISTINCT CASE
				WHEN tk.status = 'done' AND tk.updated_at >= NOW() - INTERVAL 7 DAY
				THEN tk.id
			END) AS done_last_7d
		FROM teams t
		LEFT JOIN team_members tm ON tm.team_id = t.id
		LEFT JOIN tasks tk ON tk.team_id = t.id
		GROUP BY t.id, t.name
		ORDER BY t.id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.TeamStatsRow
	for rows.Next() {
		var row domain.TeamStatsRow
		if err := rows.Scan(&row.TeamID, &row.TeamName, &row.MemberCount, &row.DoneTasksLast7d); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
