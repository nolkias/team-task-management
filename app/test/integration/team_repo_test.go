package integration

import (
	"context"
	"testing"

	"teamtask/internal/domain"
	"teamtask/internal/repository/mysql"
)

func TestTeamRepo_CreateWithOwner(t *testing.T) {
	db := setupMySQL(t)
	userRepo := mysql.NewUserRepo(db)
	teamRepo := mysql.NewTeamRepo(db)
	ctx := context.Background()

	ownerID, err := userRepo.Create(ctx, &domain.User{Email: "owner@example.com", PasswordHash: "x", Name: "Owner"})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	teamID, err := teamRepo.CreateWithOwner(ctx, &domain.Team{Name: "Engineering"}, ownerID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	membership, err := teamRepo.GetMembership(ctx, teamID, ownerID)
	if err != nil {
		t.Fatalf("expected owner membership to exist: %v", err)
	}
	if membership.Role != domain.RoleOwner {
		t.Fatalf("expected owner role, got %s", membership.Role)
	}

	count, err := teamRepo.CountDistinctMembers(ctx, teamID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 member, got %d", count)
	}
}

func TestTeamRepo_TeamStats(t *testing.T) {
	db := setupMySQL(t)
	userRepo := mysql.NewUserRepo(db)
	teamRepo := mysql.NewTeamRepo(db)
	taskRepo := mysql.NewTaskRepo(db)
	ctx := context.Background()

	owner, _ := userRepo.Create(ctx, &domain.User{Email: "owner2@example.com", PasswordHash: "x", Name: "Owner"})
	member, _ := userRepo.Create(ctx, &domain.User{Email: "member2@example.com", PasswordHash: "x", Name: "Member"})

	teamID, _ := teamRepo.CreateWithOwner(ctx, &domain.Team{Name: "Stats Team"}, owner)
	_ = teamRepo.AddMember(ctx, teamID, member, domain.RoleMember)

	taskID, err := taskRepo.Create(ctx, &domain.Task{TeamID: teamID, CreatedBy: owner, Title: "Task 1"})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	done := domain.StatusDone
	task, err := taskRepo.GetByID(ctx, taskID)
	if err != nil {
		t.Fatalf("failed to fetch task: %v", err)
	}
	task.Status = done
	if err := taskRepo.UpdateWithHistory(ctx, task, nil); err != nil {
		t.Fatalf("failed to update task: %v", err)
	}

	stats, err := teamRepo.TeamStats(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var found *domain.TeamStatsRow
	for i := range stats {
		if stats[i].TeamID == teamID {
			found = &stats[i]
		}
	}
	if found == nil {
		t.Fatalf("expected team %d in stats, got %+v", teamID, stats)
	}
	if found.MemberCount != 2 {
		t.Fatalf("expected 2 members, got %d", found.MemberCount)
	}
	if found.DoneTasksLast7d != 1 {
		t.Fatalf("expected 1 done task in last 7 days, got %d", found.DoneTasksLast7d)
	}
}
