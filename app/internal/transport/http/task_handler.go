package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"teamtask/internal/domain"
	"teamtask/internal/middleware"
	"teamtask/internal/repository"
	"teamtask/internal/service"
)

type TaskHandler struct {
	tasks           *service.TaskService
	history         repository.TaskHistoryRepository
	defaultPageSize int
	maxPageSize     int
}

func NewTaskHandler(tasks *service.TaskService, history repository.TaskHistoryRepository, defaultPageSize, maxPageSize int) *TaskHandler {
	return &TaskHandler{tasks: tasks, history: history, defaultPageSize: defaultPageSize, maxPageSize: maxPageSize}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	task := &domain.Task{
		TeamID:      req.TeamID,
		AssigneeID:  req.AssigneeID,
		CreatedBy:   userID,
		Title:       req.Title,
		Description: req.Description,
	}

	id, err := h.tasks.CreateTask(r.Context(), task)
	if err != nil {
		writeError(w, err)
		return
	}

	task.ID = id
	writeJSON(w, http.StatusCreated, toTaskResponse(*task))
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	teamID, err := strconv.ParseInt(q.Get("team_id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "team_id is required"})
		return
	}

	filter := domain.TaskFilter{TeamID: teamID}
	if status := q.Get("status"); status != "" {
		filter.Status = domain.Status(status)
	}
	if assigneeStr := q.Get("assignee_id"); assigneeStr != "" {
		assigneeID, err := strconv.ParseInt(assigneeStr, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid assignee_id"})
			return
		}
		filter.AssigneeID = &assigneeID
	}

	page := parsePositiveInt(q.Get("page"), 1)
	pageSize := parsePositiveInt(q.Get("page_size"), h.defaultPageSize)
	if pageSize > h.maxPageSize {
		pageSize = h.maxPageSize
	}
	pr := domain.PageRequest{Page: page, PageSize: pageSize}

	tasks, total, err := h.tasks.ListTasks(r.Context(), filter, pr)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := taskListResponse{Tasks: make([]taskResponse, 0, len(tasks)), Total: total, Page: page}
	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, toTaskResponse(t))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid task id"})
		return
	}

	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	update := service.TaskUpdate{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		AssigneeID:  req.AssigneeID,
	}

	task, err := h.tasks.UpdateTask(r.Context(), taskID, userID, update)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toTaskResponse(*task))
}

func (h *TaskHandler) History(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid task id"})
		return
	}

	history, err := h.history.ListByTaskID(r.Context(), taskID)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]taskHistoryResponse, 0, len(history))
	for _, hist := range history {
		resp = append(resp, taskHistoryResponse{
			ID:           hist.ID,
			TaskID:       hist.TaskID,
			ChangedBy:    hist.ChangedBy,
			FieldChanged: hist.FieldChanged,
			OldValue:     hist.OldValue,
			NewValue:     hist.NewValue,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func parsePositiveInt(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
