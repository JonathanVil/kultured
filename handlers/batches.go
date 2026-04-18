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
	DaysElapsed   int
	LatestGravity float64
	HasGravity    bool
}

type BatchDetailData struct {
	Batch     models.Batch
	Readings  []models.Reading
	Stats     BatchStats
	NextStage string
}

type BatchStats struct {
	FermentationDays int
	ABV              float64
	HasABV           bool
	SugarsRemaining  float64
	HasSugars        bool
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
		g, hasG := models.GetLatestGravityForBatch(h.DB, b.ID)
		summaries = append(summaries, BatchSummary{
			Batch:         b,
			DaysElapsed:   days,
			LatestGravity: g,
			HasGravity:    hasG,
		})
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

	readings, err := models.GetReadingsForBatch(h.DB, id)
	if err != nil {
		http.Error(w, "failed to fetch readings", http.StatusInternalServerError)
		return
	}

	days, _ := calc.FermentationDays(batch.StartedAt)
	stats := BatchStats{FermentationDays: days}

	// Find OG (oldest gravity reading) and FG (latest gravity reading)
	var og, fg float64
	var hasOG, hasFG bool
	for _, reading := range readings {
		if reading.Gravity.Valid {
			fg = reading.Gravity.Float64
			hasFG = true
			break // readings are DESC ordered, so first match is latest
		}
	}
	for i := len(readings) - 1; i >= 0; i-- {
		if readings[i].Gravity.Valid {
			og = readings[i].Gravity.Float64
			hasOG = true
			break
		}
	}

	if hasOG && hasFG && og > fg {
		stats.ABV = calc.ABV(og, fg)
		stats.HasABV = true
		stats.SugarsRemaining = calc.SugarsRemaining(batch.SugarG, og-fg)
		stats.HasSugars = true
	}

	data := BatchDetailData{
		Batch:     batch,
		Readings:  readings,
		Stats:     stats,
		NextStage: nextStage(batch.Stage),
	}

	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/batch.html",
		"templates/partials/reading_row.html",
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
