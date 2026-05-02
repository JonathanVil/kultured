package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"github.com/JonathanVil/kultured/calc"
	"github.com/JonathanVil/kultured/models"
	"github.com/go-chi/chi/v5"
)

type BatchHandler struct {
	DB *sql.DB
}

type BatchSummary struct {
	models.Batch
	DaysElapsed int
}

type BatchDetailData struct {
	Batch     models.Batch
	Notes     []models.Note
	Stats     BatchStats
	NextStage string
}

type BatchStats struct {
	FermentationDays int
}

func nextStage(current string) string {
	switch current {
	case "f1":
		return "f2"
	case "f2":
		return "bottled"
	case "bottled":
		return "done"
	default:
		return ""
	}
}

func (h *BatchHandler) Index(w http.ResponseWriter, r *http.Request) {
	batches, err := models.GetAllBatches(h.DB)
	if err != nil {
		http.Error(w, "failed to fetch batches", http.StatusInternalServerError)
		return
	}

	summaries := make([]BatchSummary, 0, len(batches))
	for _, b := range batches {
		days, _ := calc.FermentationDays(b.StartedAt)
		summaries = append(summaries, BatchSummary{Batch: b, DaysElapsed: days})
	}

	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/index.html",
		"templates/partials/batch_card.html",
	)
	if err != nil {
		http.Error(w, "failed to parse templates", http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "layout", summaries)
}

func (h *BatchHandler) New(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/new_batch.html",
	)
	if err != nil {
		http.Error(w, "failed to parse templates", http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "layout", nil)
}

func (h *BatchHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	sugarG, err := strconv.ParseFloat(r.FormValue("sugar_g"), 64)
	if err != nil {
		http.Error(w, "invalid sugar value", http.StatusBadRequest)
		return
	}

	teaVolumeL, err := strconv.ParseFloat(r.FormValue("tea_volume_l"), 64)
	if err != nil {
		http.Error(w, "invalid tea volume value", http.StatusBadRequest)
		return
	}

	scobyVolumeMl, err := strconv.ParseFloat(r.FormValue("scoby_volume_ml"), 64)
	if err != nil {
		http.Error(w, "invalid SCOBY volume value", http.StatusBadRequest)
		return
	}

	teaG, err := strconv.ParseFloat(r.FormValue("tea_g"), 64)
	if err != nil {
		http.Error(w, "invalid tea amount value", http.StatusBadRequest)
		return
	}

	steepMin, err := strconv.ParseFloat(r.FormValue("steep_min"), 64)
	if err != nil {
		http.Error(w, "invalid steep time value", http.StatusBadRequest)
		return
	}

	batch := models.Batch{
		Name:          r.FormValue("name"),
		StartedAt:     r.FormValue("started_at"),
		TeaType:       r.FormValue("tea_type"),
		TeaG:          teaG,
		SteepMin:      steepMin,
		SugarG:        sugarG,
		TeaVolumeL:    teaVolumeL,
		ScobyVolumeMl: scobyVolumeMl,
		Stage:         "f1",
	}

	if _, err := models.CreateBatch(h.DB, batch); err != nil {
		http.Error(w, "failed to create batch", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *BatchHandler) Show(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid batch id", http.StatusBadRequest)
		return
	}

	batch, err := models.GetBatch(h.DB, id)
	if err != nil {
		http.Error(w, "batch not found", http.StatusNotFound)
		return
	}

	notes, err := models.GetNotesForBatch(h.DB, id)
	if err != nil {
		http.Error(w, "failed to fetch notes", http.StatusInternalServerError)
		return
	}

	days, _ := calc.FermentationDays(batch.StartedAt)

	data := BatchDetailData{
		Batch:     batch,
		Notes:     notes,
		Stats:     BatchStats{FermentationDays: days},
		NextStage: nextStage(batch.Stage),
	}

	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/batch.html",
		"templates/partials/stats.html",
	)
	if err != nil {
		http.Error(w, "failed to parse templates", http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "layout", data)
}

func (h *BatchHandler) UpdateStage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid batch id", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	stage := r.FormValue("stage")
	validStages := map[string]bool{"f1": true, "f2": true, "bottled": true, "done": true}
	if !validStages[stage] {
		http.Error(w, "invalid stage", http.StatusBadRequest)
		return
	}

	if err := models.UpdateStage(h.DB, id, stage); err != nil {
		http.Error(w, "failed to update stage", http.StatusInternalServerError)
		return
	}

	idStr := strconv.Itoa(id)
	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/batches/"+idStr)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/batches/"+idStr, http.StatusSeeOther)
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
