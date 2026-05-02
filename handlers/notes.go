package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/JonathanVil/kultured/models"
	"github.com/go-chi/chi/v5"
)

type NoteHandler struct {
	DB *sql.DB
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	batchID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid batch id", http.StatusBadRequest)
		return
	}

	var req struct {
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Note == "" {
		http.Error(w, "note is required", http.StatusBadRequest)
		return
	}

	id, err := models.CreateNote(h.DB, models.Note{BatchID: batchID, Note: req.Note})
	if err != nil {
		http.Error(w, "failed to create note", http.StatusInternalServerError)
		return
	}

	note, err := models.GetNote(h.DB, int(id))
	if err != nil {
		http.Error(w, "failed to retrieve note", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, noteResponse{
		ID:        note.ID,
		Note:      note.Note,
		CreatedAt: note.CreatedAt,
	})
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid note id", http.StatusBadRequest)
		return
	}

	if err := models.DeleteNote(h.DB, id); err != nil {
		http.Error(w, "failed to delete note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
