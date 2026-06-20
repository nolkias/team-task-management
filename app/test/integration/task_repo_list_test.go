package integration

import (
	"context"
	"testing"

	"teamtask/internal/domain"
	"teamtask/internal/repository/mysql"
)

func TestTaskRepo_ListWithFiltersAndPagination(t *testing.T) {
	db := setupMySQL(t)
	userRepo := mysql.NewUserRepo(db)
	teamRepo := mysql.NewTeamRepo(db)
	taskRepo := mysql.NewTaskRepo(db)
	ctx := context.Background()

	owner, _ := userRepo.Create(ctx, &domain.User{Email: "lister@example.com", PasswordHash: "x", Name: "Lister"})
	teamID, _ := teamRepo.CreateWithOwner(ctx, &domain.Team{Name: "List Team"}, owner)

	for i := 0; i < 5; i++ {
		status := domain.StatusTodo
		if i%2 == 0 {
			status = domain.StatusDone
		}
		if _, err := taskRepo.Create(ctx, &domain.Task{TeamID: teamID, CreatedBy: owner, Title: "Task", Status: status}); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	doneTasks, total, err := taskRepo.List(ctx, domain.TaskFilter{TeamID: teamID, Status: domain.StatusDone}, domain.PageRequest{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 3 || len(doneTasks) != 3 {
		t.Fatalf("expected 3 done tasks, got total=%d len=%d", total, len(doneTasks))
	}

	page1, total, err := taskRepo.List(ctx, domain.TaskFilter{TeamID: teamID}, domain.PageRequest{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 || len(page1) != 2 {
		t.Fatalf("expected total=5 page-len=2, got total=%d len=%d", total, len(page1))
	}

	page3, _, err := taskRepo.List(ctx, domain.TaskFilter{TeamID: teamID}, domain.PageRequest{Page: 3, PageSize: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page3) != 1 {
		t.Fatalf("expected 1 task on last page, got %d", len(page3))
	}
}
