package service

import (
	"context"
	"fmt"

	"teamtask/internal/cache"
	"teamtask/internal/domain"
	"teamtask/internal/repository"
)

type TaskService struct {
	tasks repository.TaskRepository
	teams repository.TeamRepository
	cache *cache.TaskListCache
}

func NewTaskService(tasks repository.TaskRepository, teams repository.TeamRepository, taskListCache *cache.TaskListCache) *TaskService {
	return &TaskService{tasks: tasks, teams: teams, cache: taskListCache}
}

func (s *TaskService) CreateTask(ctx context.Context, t *domain.Task) (int64, error) {
	if _, err := s.teams.GetMembership(ctx, t.TeamID, t.CreatedBy); err != nil {
		return 0, err
	}

	if t.AssigneeID != nil {
		if _, err := s.teams.GetMembership(ctx, t.TeamID, *t.AssigneeID); err != nil {
			return 0, domain.ErrAssigneeNotMember
		}
	}

	if t.Status == "" {
		t.Status = domain.StatusTodo
	}

	id, err := s.tasks.Create(ctx, t)
	if err != nil {
		return 0, err
	}

	if s.cache != nil {
		_ = s.cache.InvalidateTeam(ctx, t.TeamID)
	}

	return id, nil
}

type TaskUpdate struct {
	Title       *string
	Description *string
	Status      *domain.Status
	AssigneeID  **int64
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID, updatedBy int64, update TaskUpdate) (*domain.Task, error) {
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if _, err := s.teams.GetMembership(ctx, t.TeamID, updatedBy); err != nil {
		return nil, err
	}

	var changes []domain.TaskHistory

	if update.Title != nil && *update.Title != t.Title {
		changes = append(changes, domain.TaskHistory{
			TaskID: taskID, ChangedBy: updatedBy, FieldChanged: "title",
			OldValue: t.Title, NewValue: *update.Title,
		})
		t.Title = *update.Title
	}

	if update.Description != nil && *update.Description != t.Description {
		changes = append(changes, domain.TaskHistory{
			TaskID: taskID, ChangedBy: updatedBy, FieldChanged: "description",
			OldValue: t.Description, NewValue: *update.Description,
		})
		t.Description = *update.Description
	}

	if update.Status != nil && *update.Status != t.Status {
		changes = append(changes, domain.TaskHistory{
			TaskID: taskID, ChangedBy: updatedBy, FieldChanged: "status",
			OldValue: string(t.Status), NewValue: string(*update.Status),
		})
		t.Status = *update.Status
	}

	if update.AssigneeID != nil && !sameAssignee(*update.AssigneeID, t.AssigneeID) {
		newAssignee := *update.AssigneeID
		if newAssignee != nil {
			if _, err := s.teams.GetMembership(ctx, t.TeamID, *newAssignee); err != nil {
				return nil, domain.ErrAssigneeNotMember
			}
		}
		changes = append(changes, domain.TaskHistory{
			TaskID: taskID, ChangedBy: updatedBy, FieldChanged: "assignee_id",
			OldValue: formatAssignee(t.AssigneeID), NewValue: formatAssignee(newAssignee),
		})
		t.AssigneeID = newAssignee
	}

	if len(changes) == 0 {
		return t, nil
	}

	if err := s.tasks.UpdateWithHistory(ctx, t, changes); err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.InvalidateTeam(ctx, t.TeamID)
	}

	return t, nil
}

func (s *TaskService) ListTasks(ctx context.Context, f domain.TaskFilter, p domain.PageRequest) ([]domain.Task, int, error) {
	if s.cache != nil {
		key := cache.BuildTaskListKey(f, p)
		if cached, err := s.cache.Get(ctx, key); err == nil && cached != nil {
			return cached.Tasks, cached.Total, nil
		}
	}

	tasks, total, err := s.tasks.List(ctx, f, p)
	if err != nil {
		return nil, 0, err
	}

	if s.cache != nil {
		key := cache.BuildTaskListKey(f, p)
		_ = s.cache.Set(ctx, f.TeamID, key, &cache.TaskListResult{Tasks: tasks, Total: total})
	}

	return tasks, total, nil
}

func sameAssignee(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func formatAssignee(id *int64) string {
	if id == nil {
		return ""
	}
	return fmt.Sprintf("%d", *id)
}
