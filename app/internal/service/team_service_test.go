package service

import (
	"context"
	"errors"
	"testing"

	"teamtask/internal/domain"
)

type stubTeamRepo struct {
	teams       map[int64]*domain.Team
	memberships map[int64]map[int64]domain.Role
	nextID      int64
}

func newStubTeamRepo() *stubTeamRepo {
	return &stubTeamRepo{
		teams:       map[int64]*domain.Team{},
		memberships: map[int64]map[int64]domain.Role{},
	}
}

func (s *stubTeamRepo) CreateWithOwner(ctx context.Context, team *domain.Team, ownerID int64) (int64, error) {
	s.nextID++
	team.ID = s.nextID
	s.teams[team.ID] = team
	s.memberships[team.ID] = map[int64]domain.Role{ownerID: domain.RoleOwner}
	return team.ID, nil
}

func (s *stubTeamRepo) ListForUser(ctx context.Context, userID int64) ([]domain.Team, error) {
	var result []domain.Team
	for teamID, members := range s.memberships {
		if _, ok := members[userID]; ok {
			result = append(result, *s.teams[teamID])
		}
	}
	return result, nil
}

func (s *stubTeamRepo) GetMembership(ctx context.Context, teamID, userID int64) (*domain.TeamMember, error) {
	members, ok := s.memberships[teamID]
	if !ok {
		return nil, domain.ErrNotTeamMember
	}
	role, ok := members[userID]
	if !ok {
		return nil, domain.ErrNotTeamMember
	}
	return &domain.TeamMember{TeamID: teamID, UserID: userID, Role: role}, nil
}

func (s *stubTeamRepo) AddMember(ctx context.Context, teamID, userID int64, role domain.Role) error {
	if s.memberships[teamID] == nil {
		s.memberships[teamID] = map[int64]domain.Role{}
	}
	s.memberships[teamID][userID] = role
	return nil
}

func (s *stubTeamRepo) CountDistinctMembers(ctx context.Context, teamID int64) (int, error) {
	return len(s.memberships[teamID]), nil
}

func (s *stubTeamRepo) TeamStats(ctx context.Context) ([]domain.TeamStatsRow, error) {
	return nil, nil
}

type stubEmailService struct {
	shouldFail bool
	calls      int
}

func (s *stubEmailService) SendInvite(ctx context.Context, toEmail, teamName string) error {
	s.calls++
	if s.shouldFail {
		return ErrEmailDeliveryFailed
	}
	return nil
}

func TestTeamService_CreateTeam(t *testing.T) {
	teams := newStubTeamRepo()
	svc := NewTeamService(teams, newStubUserRepo(), &stubEmailService{})

	id, err := svc.CreateTeam(context.Background(), "Engineering", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	membership, err := teams.GetMembership(context.Background(), id, 1)
	if err != nil {
		t.Fatalf("expected owner membership, got error: %v", err)
	}
	if membership.Role != domain.RoleOwner {
		t.Fatalf("expected owner role, got %s", membership.Role)
	}
}

func TestTeamService_ListTeams(t *testing.T) {
	teams := newStubTeamRepo()
	svc := NewTeamService(teams, newStubUserRepo(), &stubEmailService{})

	id1, _ := svc.CreateTeam(context.Background(), "Engineering", 1)
	_, _ = svc.CreateTeam(context.Background(), "Marketing", 2)

	result, err := svc.ListTeams(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].ID != id1 {
		t.Fatalf("expected only the team user 1 owns, got %+v", result)
	}
}

func TestTeamService_InviteMember(t *testing.T) {
	tests := []struct {
		name          string
		inviterRole   domain.Role
		emailFails    bool
		alreadyMember bool
		inviteeExists bool
		wantErr       error
	}{
		{name: "owner can invite", inviterRole: domain.RoleOwner, inviteeExists: true},
		{name: "member cannot invite", inviterRole: domain.RoleMember, inviteeExists: true, wantErr: domain.ErrInsufficientRole},
		{name: "invitee not registered", inviterRole: domain.RoleOwner, inviteeExists: false, wantErr: domain.ErrUserNotFound},
		{name: "invitee already member", inviterRole: domain.RoleOwner, inviteeExists: true, alreadyMember: true, wantErr: domain.ErrAlreadyTeamMember},
		{name: "email service fails", inviterRole: domain.RoleOwner, inviteeExists: true, emailFails: true, wantErr: ErrEmailDeliveryFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teams := newStubTeamRepo()
			users := newStubUserRepo()
			email := &stubEmailService{shouldFail: tt.emailFails}
			svc := NewTeamService(teams, users, email)

			teamID := int64(1)
			teams.teams[teamID] = &domain.Team{ID: teamID, Name: "Engineering"}
			teams.memberships[teamID] = map[int64]domain.Role{1: tt.inviterRole}

			if tt.inviteeExists {
				users.byEmail["invitee@example.com"] = &domain.User{ID: 2, Email: "invitee@example.com"}
				users.byID[2] = users.byEmail["invitee@example.com"]
			}
			if tt.alreadyMember {
				teams.memberships[teamID][2] = domain.RoleMember
			}

			err := svc.InviteMember(context.Background(), teamID, 1, "invitee@example.com")

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if _, err := teams.GetMembership(context.Background(), teamID, 2); err != nil {
				t.Fatalf("expected invitee to become a member: %v", err)
			}
			if email.calls != 1 {
				t.Fatalf("expected email service to be called once, got %d", email.calls)
			}
		})
	}
}
