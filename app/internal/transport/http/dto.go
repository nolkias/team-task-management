package http

import "teamtask/internal/domain"

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type createTeamRequest struct {
	Name string `json:"name"`
}

type teamResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedBy int64  `json:"created_by"`
}

type inviteRequest struct {
	Email string `json:"email"`
}

type createTaskRequest struct {
	TeamID      int64  `json:"team_id"`
	AssigneeID  *int64 `json:"assignee_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateTaskRequest struct {
	Title       *string        `json:"title"`
	Description *string        `json:"description"`
	Status      *domain.Status `json:"status"`
	AssigneeID  **int64        `json:"assignee_id"`
}

type taskResponse struct {
	ID          int64         `json:"id"`
	TeamID      int64         `json:"team_id"`
	AssigneeID  *int64        `json:"assignee_id"`
	CreatedBy   int64         `json:"created_by"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Status      domain.Status `json:"status"`
}

func toTaskResponse(t domain.Task) taskResponse {
	return taskResponse{
		ID:          t.ID,
		TeamID:      t.TeamID,
		AssigneeID:  t.AssigneeID,
		CreatedBy:   t.CreatedBy,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
	}
}

type taskListResponse struct {
	Tasks []taskResponse `json:"tasks"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
}

type taskHistoryResponse struct {
	ID           int64  `json:"id"`
	TaskID       int64  `json:"task_id"`
	ChangedBy    int64  `json:"changed_by"`
	FieldChanged string `json:"field_changed"`
	OldValue     string `json:"old_value"`
	NewValue     string `json:"new_value"`
}
