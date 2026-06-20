package http

import (
	"net/http"

	"teamtask/internal/repository"
)

type AdminHandler struct {
	tasks repository.TaskRepository
	teams repository.TeamRepository
}

func NewAdminHandler(tasks repository.TaskRepository, teams repository.TeamRepository) *AdminHandler {
	return &AdminHandler{tasks: tasks, teams: teams}
}

func (h *AdminHandler) OrphanedAssignees(w http.ResponseWriter, r *http.Request) {
	rows, err := h.tasks.OrphanedAssignees(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (h *AdminHandler) TopCreators(w http.ResponseWriter, r *http.Request) {
	rows, err := h.tasks.TopCreatorsPerTeamThisMonth(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (h *AdminHandler) TeamStats(w http.ResponseWriter, r *http.Request) {
	rows, err := h.teams.TeamStats(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
