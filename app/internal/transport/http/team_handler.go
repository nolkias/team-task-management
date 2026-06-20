package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"teamtask/internal/middleware"
	"teamtask/internal/service"
)

type TeamHandler struct {
	teams *service.TeamService
}

func NewTeamHandler(teams *service.TeamService) *TeamHandler {
	return &TeamHandler{teams: teams}
}

func (h *TeamHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req createTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	id, err := h.teams.CreateTeam(r.Context(), req.Name, userID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, teamResponse{ID: id, Name: req.Name, CreatedBy: userID})
}

func (h *TeamHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	teams, err := h.teams.ListTeams(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]teamResponse, 0, len(teams))
	for _, t := range teams {
		resp = append(resp, teamResponse{ID: t.ID, Name: t.Name, CreatedBy: t.CreatedBy})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TeamHandler) Invite(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	teamID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid team id"})
		return
	}

	var req inviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.teams.InviteMember(r.Context(), teamID, userID, req.Email); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "invited"})
}
