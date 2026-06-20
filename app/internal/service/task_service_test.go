package service

import (
	"context"
	"errors"
	"testing"

	"teamtask/internal/domain"
)

type stubTaskRepo struct {
	tasks     map[int64]*domain.Task
	history   []domain.TaskHistory
	nextID    int64
	updateErr error
}

func newStubTaskRepo() *stubTaskRepo {
	return &stubTaskRepo{tasks: map[int64]*domain.Task{}}
}

func (s *stubTaskRepo) Create(ctx context.Context, t *domain.Task) (int64, error) {
	s.nextID++
	t.ID = s.nextID
	stored := *t
	s.tasks[t.ID] = &stored
	return t.ID, nil
}

func (s *stubTaskRepo) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	t, ok := s.tasks[id]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}
	copy := *t
	return &copy, nil
}

func (s *stubTaskRepo) UpdateWithHistory(ctx context.Context, t *domain.Task, changes []domain.TaskHistory) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	stored := *t
	s.tasks[t.ID] = &stored
	s.history = append(s.history, changes...)
	return nil
}

func (s *stubTaskRepo) List(ctx context.Context, f domain.TaskFilter, p domain.PageRequest) ([]domain.Task, int, error) {
	var result []domain.Task
	for _, t := range s.tasks {
		if t.TeamID == f.TeamID {
			result = append(result, *t)
		}
	}
	return result, len(result), nil
}

func (s *stubTaskRepo) TopCreatorsPerTeamThisMonth(ctx context.Context) ([]domain.TopCreatorRow, error) {
	return nil, nil
}

func (s *stubTaskRepo) OrphanedAssignees(ctx context.Context) ([]domain.OrphanedAssigneeRow, error) {
	return nil, nil
}

func ptr64(v int64) *int64 { return &v }

func TestTaskService_CreateTask(t *testing.T) {
	tests := []struct {
		name             string
		creatorRole      domain.Role
		assigneeID       *int64
		assigneeIsMember bool
		wantErr          error
	}{
		{name: "creator is member, no assignee", creatorRole: domain.RoleMember},
		{name: "creator not a member", wantErr: domain.ErrNotTeamMember},
		{name: "assignee not a member", creatorRole: domain.RoleMember, assigneeID: ptr64(2), assigneeIsMember: false, wantErr: domain.ErrAssigneeNotMember},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teams := newStubTeamRepo()
			tasks := newStubTaskRepo()
			svc := NewTaskService(tasks, teams, nil)

			teamID := int64(1)
			teams.memberships[teamID] = map[int64]domain.Role{}
			if tt.creatorRole != "" {
				teams.memberships[teamID][1] = tt.creatorRole
			}
			if tt.assigneeID != nil && tt.assigneeIsMember {
				teams.memberships[teamID][*tt.assigneeID] = domain.RoleMember
			}

			task := &domain.Task{TeamID: teamID, CreatedBy: 1, Title: "Test task", AssigneeID: tt.assigneeID}
			_, err := svc.CreateTask(context.Background(), task)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestTaskService_UpdateTask(t *testing.T) {
	t.Run("status change is recorded in history", func(t *testing.T) {
		teams := newStubTeamRepo()
		tasks := newStubTaskRepo()
		svc := NewTaskService(tasks, teams, nil)

		teamID := int64(1)
		teams.memberships[teamID] = map[int64]domain.Role{1: domain.RoleMember}
		tasks.tasks[1] = &domain.Task{ID: 1, TeamID: teamID, Title: "Task", Status: domain.StatusTodo}

		newStatus := domain.StatusDone
		updated, err := svc.UpdateTask(context.Background(), 1, 1, TaskUpdate{Status: &newStatus})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != domain.StatusDone {
			t.Fatalf("expected status done, got %s", updated.Status)
		}
		if len(tasks.history) != 1 || tasks.history[0].FieldChanged != "status" {
			t.Fatalf("expected one status history entry, got %+v", tasks.history)
		}
	})

	t.Run("non-member cannot update", func(t *testing.T) {
		teams := newStubTeamRepo()
		tasks := newStubTaskRepo()
		svc := NewTaskService(tasks, teams, nil)

		teamID := int64(1)
		teams.memberships[teamID] = map[int64]domain.Role{}
		tasks.tasks[1] = &domain.Task{ID: 1, TeamID: teamID, Title: "Task", Status: domain.StatusTodo}

		newStatus := domain.StatusDone
		_, err := svc.UpdateTask(context.Background(), 1, 1, TaskUpdate{Status: &newStatus})
		if !errors.Is(err, domain.ErrNotTeamMember) {
			t.Fatalf("expected ErrNotTeamMember, got %v", err)
		}
	})

	t.Run("reassign to non-member is rejected", func(t *testing.T) {
		teams := newStubTeamRepo()
		tasks := newStubTaskRepo()
		svc := NewTaskService(tasks, teams, nil)

		teamID := int64(1)
		teams.memberships[teamID] = map[int64]domain.Role{1: domain.RoleMember}
		tasks.tasks[1] = &domain.Task{ID: 1, TeamID: teamID, Title: "Task", Status: domain.StatusTodo}

		newAssignee := ptr64(99)
		_, err := svc.UpdateTask(context.Background(), 1, 1, TaskUpdate{AssigneeID: &newAssignee})
		if !errors.Is(err, domain.ErrAssigneeNotMember) {
			t.Fatalf("expected ErrAssigneeNotMember, got %v", err)
		}
	})

	t.Run("reassign to a valid member is recorded", func(t *testing.T) {
		teams := newStubTeamRepo()
		tasks := newStubTaskRepo()
		svc := NewTaskService(tasks, teams, nil)

		teamID := int64(1)
		teams.memberships[teamID] = map[int64]domain.Role{1: domain.RoleMember, 2: domain.RoleMember}
		tasks.tasks[1] = &domain.Task{ID: 1, TeamID: teamID, Title: "Task", Status: domain.StatusTodo}

		newAssignee := ptr64(2)
		updated, err := svc.UpdateTask(context.Background(), 1, 1, TaskUpdate{AssigneeID: &newAssignee})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.AssigneeID == nil || *updated.AssigneeID != 2 {
			t.Fatalf("expected assignee 2, got %+v", updated.AssigneeID)
		}
	})

	t.Run("title and description changes are recorded", func(t *testing.T) {
		teams := newStubTeamRepo()
		tasks := newStubTaskRepo()
		svc := NewTaskService(tasks, teams, nil)

		teamID := int64(1)
		teams.memberships[teamID] = map[int64]domain.Role{1: domain.RoleMember}
		tasks.tasks[1] = &domain.Task{ID: 1, TeamID: teamID, Title: "Old title", Description: "Old desc", Status: domain.StatusTodo}

		newTitle := "New title"
		newDesc := "New desc"
		updated, err := svc.UpdateTask(context.Background(), 1, 1, TaskUpdate{Title: &newTitle, Description: &newDesc})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Title != newTitle || updated.Description != newDesc {
			t.Fatalf("expected title/description to be updated, got %+v", updated)
		}
		if len(tasks.history) != 2 {
			t.Fatalf("expected 2 history entries, got %+v", tasks.history)
		}
	})

	t.Run("no-op update when values unchanged", func(t *testing.T) {
		teams := newStubTeamRepo()
		tasks := newStubTaskRepo()
		svc := NewTaskService(tasks, teams, nil)

		teamID := int64(1)
		teams.memberships[teamID] = map[int64]domain.Role{1: domain.RoleMember}
		tasks.tasks[1] = &domain.Task{ID: 1, TeamID: teamID, Title: "Task", Status: domain.StatusTodo}

		sameStatus := domain.StatusTodo
		_, err := svc.UpdateTask(context.Background(), 1, 1, TaskUpdate{Status: &sameStatus})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks.history) != 0 {
			t.Fatalf("expected no history entries for unchanged status, got %+v", tasks.history)
		}
	})
}

func TestTaskService_ListTasks(t *testing.T) {
	teams := newStubTeamRepo()
	tasks := newStubTaskRepo()
	svc := NewTaskService(tasks, teams, nil)

	teamID := int64(1)
	tasks.tasks[1] = &domain.Task{ID: 1, TeamID: teamID, Title: "A"}
	tasks.tasks[2] = &domain.Task{ID: 2, TeamID: 2, Title: "B"}

	result, total, err := svc.ListTasks(context.Background(), domain.TaskFilter{TeamID: teamID}, domain.PageRequest{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(result) != 1 {
		t.Fatalf("expected 1 task for team %d, got %d (total=%d)", teamID, len(result), total)
	}
}
