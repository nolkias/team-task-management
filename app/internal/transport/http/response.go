package http

import (
	"encoding/json"
	"net/http"

	"teamtask/internal/apperror"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	status := apperror.StatusFor(err)
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
