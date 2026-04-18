package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"github.com/JonathanVil/kultured/models"

	"github.com/go-chi/chi/v5"
)

type BatchHandler struct {
	DB *sql.DB
}

func (h *BatchHandler) Index(w http.ResponseWriter, r *http.Request) {
	batches, err := models.GetAllBatches(h.DB)
	if err != nil {
		http.Error(w, "failed to fetch batches", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/index.html",
	)
	if err != nil {
		http.Error(w, "failed to parse templates", http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "layout", batches)
}

func (h *BatchHandler) New(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/new_batch.html",
	)
	if err != nil {
		http.Error(w, "failed to parse templates", http.StatusInternalServerError)
	}

	tmpl.ExecuteTemplate(w, "layout", nil)
}

func (h *BatchHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
	}

	sugarG, err := strconv.ParseFloat(r.FormValue("sugar_g"), 64)
	if err != nil {
		http.Error(w, "invalid sugar value", http.StatusBadRequest)
	}

	volumeL, err := strconv.ParseFloat(r.FormValue("volume_l"), 64)
	if err != nil {
		http.Error(w, "invalid volume value", http.StatusBadRequest)
		return
	}

	scobyWeightG, err := strconv.ParseFloat(r.FormValue("scoby_weight_g"), 64)
	if err != nil {
		http.Error(w, "invalid SCOBY weight value", http.StatusBadRequest)
		return
	}

	notes := r.FormValue("notes")
	batch := models.Batch{
		Name:         r.FormValue("name"),
		StartedAt:    r.FormValue("started_at"),
		TeaType:      r.FormValue("tea_type"),
		SugarG:       sugarG,
		VolumeL:      volumeL,
		ScobyWeightG: scobyWeightG,
		Stage:        "f1",
		Notes:        sql.NullString{String: notes, Valid: notes != ""},
	}

	if _, err := models.CreateBatch(h.DB, batch); err != nil {
		http.Error(w, "failed to create batch", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *BatchHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid batch id", http.StatusBadRequest)
		return
	}

	if err := models.DeleteBatch(h.DB, id); err != nil {
		http.Error(w, "failed to delete batch", http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func isHTMX(r *http.Request) bool {
    return r.Header.Get("HX-Request") == "true"
}
