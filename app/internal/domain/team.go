package domain

import (
	"errors"
	"time"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

var (
	ErrTeamNotFound      = errors.New("team not found")
	ErrNotTeamMember     = errors.New("user is not a member of this team")
	ErrInsufficientRole  = errors.New("insufficient role for this action")
	ErrAlreadyTeamMember = errors.New("user is already a member of this team")
)

type Team struct {
	ID        int64
	Name      string
	CreatedBy int64
	CreatedAt time.Time
}

type TeamMember struct {
	TeamID   int64
	UserID   int64
	Role     Role
	JoinedAt time.Time
}

type TeamStatsRow struct {
	TeamID          int64
	TeamName        string
	MemberCount     int
	DoneTasksLast7d int
}
