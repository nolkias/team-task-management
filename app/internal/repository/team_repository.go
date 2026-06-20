package repository

import (
	"context"

	"teamtask/internal/domain"
)

type TeamRepository interface {
	CreateWithOwner(ctx context.Context, team *domain.Team, ownerID int64) (int64, error)
	ListForUser(ctx context.Context, userID int64) ([]domain.Team, error)
	GetMembership(ctx context.Context, teamID, userID int64) (*domain.TeamMember, error)
	AddMember(ctx context.Context, teamID, userID int64, role domain.Role) error
	CountDistinctMembers(ctx context.Context, teamID int64) (int, error)
	TeamStats(ctx context.Context) ([]domain.TeamStatsRow, error)
}
