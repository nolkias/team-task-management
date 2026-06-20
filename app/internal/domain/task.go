package domain

import (
	"errors"
	"time"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrAssigneeNotMember = errors.New("assignee is not a member of the task's team")
	ErrInvalidStatus     = errors.New("invalid task status")
)

type Task struct {
	ID          int64
	TeamID      int64
	AssigneeID  *int64
	CreatedBy   int64
	Title       string
	Description string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskFilter struct {
	TeamID     int64
	Status     Status
	AssigneeID *int64
}

type PageRequest struct {
	Page     int
	PageSize int
}

func (p PageRequest) Offset() int {
	return (p.Page - 1) * p.PageSize
}

type TopCreatorRow struct {
	TeamID       int64
	UserID       int64
	UserName     string
	TasksCreated int
	Rank         int
}

type OrphanedAssigneeRow struct {
	TaskID     int64
	TaskTitle  string
	TeamID     int64
	AssigneeID int64
}
