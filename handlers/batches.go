package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/JonathanVil/kultured/calc"
	"github.com/JonathanVil/kultured/models"
	"github.com/go-chi/chi/v5"
)

type BatchHandler struct {
	DB *sql.DB
}

type batchResponse struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	TeaType       string   `json:"tea_type"`
	TeaG          float64  `json:"tea_g"`
	SteepMin      float64  `json:"steep_min"`
	SugarG        float64  `json:"sugar_g"`
	TeaVolumeMl   float64  `json:"tea_volume_ml"`
	ScobyVolumeMl float64  `json:"scoby_volume_ml"`
	TotalVolumeMl float64  `json:"total_volume_ml"`
	Stage         string   `json:"stage"`
	StartedAt     string   `json:"started_at"`
	StartF2       *string  `json:"start_f2"`
	DoneAt        *string  `json:"done_at"`
	CreatedAt     string   `json:"created_at"`
	F1Days        int      `json:"f1_days"`
	F2Days        int      `json:"f2_days"`
	BackslopPct   float64  `json:"backslop_pct"`
	SugarPct      float64  `json:"sugar_pct"`
	TeaGPerL      float64  `json:"tea_g_per_l"`
}

type batchDetailResponse struct {
	batchResponse
	Notes []noteResponse `json:"notes"`
}

type noteResponse struct {
	ID        int    `json:"id"`
	Note      string `json:"note"`
	CreatedAt string `json:"created_at"`
}

func toBatchResponse(b models.Batch) batchResponse {
	totalVolumeMl := b.TeaVolumeMl + b.ScobyVolumeMl

	var startF2, doneAt *string
	if b.StartF2.Valid {
		s := b.StartF2.String
		startF2 = &s
	}
	if b.DoneAt.Valid {
		s := b.DoneAt.String
		doneAt = &s
	}

	f1Days := 0
	if b.StartF2.Valid {
		f1Days, _ = calc.DaysBetween(b.StartedAt, b.StartF2.String)
	} else {
		f1Days, _ = calc.FermentationDays(b.StartedAt)
	}

	f2Days := 0
	if b.StartF2.Valid {
		if b.DoneAt.Valid {
			f2Days, _ = calc.DaysBetween(b.StartF2.String, b.DoneAt.String)
		} else {
			f2Days, _ = calc.DaysSince(b.StartF2.String)
		}
	}

	var backslopPct, sugarPct float64
	if totalVolumeMl > 0 {
		backslopPct = b.ScobyVolumeMl / totalVolumeMl * 100
		sugarPct = b.SugarG / totalVolumeMl * 100
	}

	var teaGPerL float64
	if b.TeaVolumeMl > 0 {
		teaGPerL = b.TeaG / b.TeaVolumeMl * 1000
	}

	return batchResponse{
		ID:            b.ID,
		Name:          b.Name,
		TeaType:       b.TeaType,
		TeaG:          b.TeaG,
		SteepMin:      b.SteepMin,
		SugarG:        b.SugarG,
		TeaVolumeMl:   b.TeaVolumeMl,
		ScobyVolumeMl: b.ScobyVolumeMl,
		TotalVolumeMl: totalVolumeMl,
		Stage:         b.Stage,
		StartedAt:     b.StartedAt,
		StartF2:       startF2,
		DoneAt:        doneAt,
		CreatedAt:     b.CreatedAt,
		F1Days:        f1Days,
		F2Days:        f2Days,
		BackslopPct:   backslopPct,
		SugarPct:      sugarPct,
		TeaGPerL:      teaGPerL,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *BatchHandler) List(w http.ResponseWriter, r *http.Request) {
	batches, err := models.GetAllBatches(h.DB)
	if err != nil {
		http.Error(w, "failed to fetch batches", http.StatusInternalServerError)
		return
	}
	resp := make([]batchResponse, len(batches))
	for i, b := range batches {
		resp[i] = toBatchResponse(b)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *BatchHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	noteResps := make([]noteResponse, len(notes))
	for i, n := range notes {
		noteResps[i] = noteResponse{ID: n.ID, Note: n.Note, CreatedAt: n.CreatedAt}
	}

	writeJSON(w, http.StatusOK, batchDetailResponse{
		batchResponse: toBatchResponse(batch),
		Notes:         noteResps,
	})
}

func (h *BatchHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string  `json:"name"`
		TeaType       string  `json:"tea_type"`
		TeaG          float64 `json:"tea_g"`
		SteepMin      float64 `json:"steep_min"`
		SugarG        float64 `json:"sugar_g"`
		TeaVolumeMl   float64 `json:"tea_volume_ml"`
		ScobyVolumeMl float64 `json:"scoby_volume_ml"`
		StartedAt     string  `json:"started_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.TeaType == "" || req.StartedAt == "" {
		http.Error(w, "name, tea_type, and started_at are required", http.StatusBadRequest)
		return
	}

	batch := models.Batch{
		Name:          req.Name,
		StartedAt:     req.StartedAt,
		TeaType:       req.TeaType,
		TeaG:          req.TeaG,
		SteepMin:      req.SteepMin,
		SugarG:        req.SugarG,
		TeaVolumeMl:   req.TeaVolumeMl,
		ScobyVolumeMl: req.ScobyVolumeMl,
		Stage:         "f1",
	}
	id, err := models.CreateBatch(h.DB, batch)
	if err != nil {
		http.Error(w, "failed to create batch", http.StatusInternalServerError)
		return
	}

	created, err := models.GetBatch(h.DB, int(id))
	if err != nil {
		http.Error(w, "failed to retrieve created batch", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, toBatchResponse(created))
}

func (h *BatchHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid batch id", http.StatusBadRequest)
		return
	}

	var req struct {
		Batch struct {
			Name          string  `json:"name"`
			TeaType       string  `json:"tea_type"`
			TeaG          float64 `json:"tea_g"`
			SteepMin      float64 `json:"steep_min"`
			SugarG        float64 `json:"sugar_g"`
			TeaVolumeMl   float64 `json:"tea_volume_ml"`
			ScobyVolumeMl float64 `json:"scoby_volume_ml"`
			StartedAt     string  `json:"started_at"`
		} `json:"batch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	b := models.Batch{
		ID:            id,
		Name:          req.Batch.Name,
		TeaType:       req.Batch.TeaType,
		TeaG:          req.Batch.TeaG,
		SteepMin:      req.Batch.SteepMin,
		SugarG:        req.Batch.SugarG,
		TeaVolumeMl:   req.Batch.TeaVolumeMl,
		ScobyVolumeMl: req.Batch.ScobyVolumeMl,
		StartedAt:     req.Batch.StartedAt,
	}
	if err := models.UpdateBatch(h.DB, b); err != nil {
		http.Error(w, "failed to update batch", http.StatusInternalServerError)
		return
	}

	updated, err := models.GetBatch(h.DB, id)
	if err != nil {
		http.Error(w, "failed to retrieve updated batch", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, toBatchResponse(updated))
}

func (h *BatchHandler) UpdateStage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid batch id", http.StatusBadRequest)
		return
	}

	var req struct {
		Stage string `json:"stage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	validStages := map[string]bool{"f1": true, "f2": true, "bottled": true, "done": true}
	if !validStages[req.Stage] {
		http.Error(w, "invalid stage", http.StatusBadRequest)
		return
	}

	if err := models.UpdateStage(h.DB, id, req.Stage); err != nil {
		http.Error(w, "failed to update stage", http.StatusInternalServerError)
		return
	}

	updated, err := models.GetBatch(h.DB, id)
	if err != nil {
		http.Error(w, "failed to retrieve updated batch", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, toBatchResponse(updated))
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

	w.WriteHeader(http.StatusNoContent)
}
