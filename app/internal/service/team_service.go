package service

import (
	"context"
	"errors"

	"teamtask/internal/domain"
	"teamtask/internal/repository"
)

type TeamService struct {
	teams repository.TeamRepository
	users repository.UserRepository
	email EmailService
}

func NewTeamService(teams repository.TeamRepository, users repository.UserRepository, email EmailService) *TeamService {
	return &TeamService{teams: teams, users: users, email: email}
}

func (s *TeamService) CreateTeam(ctx context.Context, name string, ownerID int64) (int64, error) {
	team := &domain.Team{Name: name}
	return s.teams.CreateWithOwner(ctx, team, ownerID)
}

func (s *TeamService) ListTeams(ctx context.Context, userID int64) ([]domain.Team, error) {
	return s.teams.ListForUser(ctx, userID)
}

func (s *TeamService) InviteMember(ctx context.Context, teamID, inviterID int64, inviteeEmail string) error {
	inviter, err := s.teams.GetMembership(ctx, teamID, inviterID)
	if err != nil {
		return err
	}
	if inviter.Role != domain.RoleOwner && inviter.Role != domain.RoleAdmin {
		return domain.ErrInsufficientRole
	}

	invitee, err := s.users.GetByEmail(ctx, inviteeEmail)
	if err != nil {
		return err
	}

	if _, err := s.teams.GetMembership(ctx, teamID, invitee.ID); err == nil {
		return domain.ErrAlreadyTeamMember
	} else if !errors.Is(err, domain.ErrNotTeamMember) {
		return err
	}

	if err := s.email.SendInvite(ctx, invitee.Email, ""); err != nil {
		return err
	}

	return s.teams.AddMember(ctx, teamID, invitee.ID, domain.RoleMember)
}
