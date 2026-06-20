package integration

import (
	"context"
	"testing"

	"teamtask/internal/domain"
	"teamtask/internal/repository/mysql"
)

func TestTaskRepo_TopCreatorsPerTeamThisMonth(t *testing.T) {
	db := setupMySQL(t)
	userRepo := mysql.NewUserRepo(db)
	teamRepo := mysql.NewTeamRepo(db)
	taskRepo := mysql.NewTaskRepo(db)
	ctx := context.Background()

	owner, _ := userRepo.Create(ctx, &domain.User{Email: "creator1@example.com", PasswordHash: "x", Name: "Creator One"})
	other, _ := userRepo.Create(ctx, &domain.User{Email: "creator2@example.com", PasswordHash: "x", Name: "Creator Two"})
	teamID, _ := teamRepo.CreateWithOwner(ctx, &domain.Team{Name: "Top Creators Team"}, owner)
	_ = teamRepo.AddMember(ctx, teamID, other, domain.RoleMember)

	for i := 0; i < 3; i++ {
		if _, err := taskRepo.Create(ctx, &domain.Task{TeamID: teamID, CreatedBy: owner, Title: "Owner task"}); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}
	if _, err := taskRepo.Create(ctx, &domain.Task{TeamID: teamID, CreatedBy: other, Title: "Other task"}); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	rows, err := taskRepo.TopCreatorsPerTeamThisMonth(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var top *domain.TopCreatorRow
	for i := range rows {
		if rows[i].TeamID == teamID && rows[i].Rank == 1 {
			top = &rows[i]
		}
	}
	if top == nil {
		t.Fatalf("expected a rank-1 creator for team %d, got %+v", teamID, rows)
	}
	if top.UserID != owner || top.TasksCreated != 3 {
		t.Fatalf("expected owner to be top creator with 3 tasks, got %+v", top)
	}
}

func TestTaskRepo_OrphanedAssignees(t *testing.T) {
	db := setupMySQL(t)
	userRepo := mysql.NewUserRepo(db)
	teamRepo := mysql.NewTeamRepo(db)
	taskRepo := mysql.NewTaskRepo(db)
	ctx := context.Background()

	owner, _ := userRepo.Create(ctx, &domain.User{Email: "owner3@example.com", PasswordHash: "x", Name: "Owner"})
	outsider, _ := userRepo.Create(ctx, &domain.User{Email: "outsider@example.com", PasswordHash: "x", Name: "Outsider"})
	teamID, _ := teamRepo.CreateWithOwner(ctx, &domain.Team{Name: "Orphan Team"}, owner)

	taskID, err := taskRepo.Create(ctx, &domain.Task{TeamID: teamID, CreatedBy: owner, Title: "Orphaned task"})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	task, err := taskRepo.GetByID(ctx, taskID)
	if err != nil {
		t.Fatalf("failed to fetch task: %v", err)
	}
	task.AssigneeID = &outsider
	if err := taskRepo.UpdateWithHistory(ctx, task, nil); err != nil {
		t.Fatalf("failed to assign outsider directly via repository: %v", err)
	}

	rows, err := taskRepo.OrphanedAssignees(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var found bool
	for _, row := range rows {
		if row.TaskID == taskID && row.AssigneeID == outsider {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected task %d with non-member assignee %d to be flagged, got %+v", taskID, outsider, rows)
	}
}

func TestTaskHistoryRepo_ListByTaskID(t *testing.T) {
	db := setupMySQL(t)
	userRepo := mysql.NewUserRepo(db)
	teamRepo := mysql.NewTeamRepo(db)
	taskRepo := mysql.NewTaskRepo(db)
	historyRepo := mysql.NewTaskHistoryRepo(db)
	ctx := context.Background()

	owner, _ := userRepo.Create(ctx, &domain.User{Email: "owner4@example.com", PasswordHash: "x", Name: "Owner"})
	teamID, _ := teamRepo.CreateWithOwner(ctx, &domain.Team{Name: "History Team"}, owner)

	taskID, err := taskRepo.Create(ctx, &domain.Task{TeamID: teamID, CreatedBy: owner, Title: "Task", Status: domain.StatusTodo})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	task, _ := taskRepo.GetByID(ctx, taskID)
	task.Status = domain.StatusDone
	changes := []domain.TaskHistory{
		{TaskID: taskID, ChangedBy: owner, FieldChanged: "status", OldValue: "todo", NewValue: "done"},
	}
	if err := taskRepo.UpdateWithHistory(ctx, task, changes); err != nil {
		t.Fatalf("failed to update with history: %v", err)
	}

	history, err := historyRepo.ListByTaskID(ctx, taskID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(history))
	}
	if history[0].FieldChanged != "status" || history[0].NewValue != "done" {
		t.Fatalf("unexpected history entry: %+v", history[0])
	}
}
