package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/JonathanVil/kultured/models"
	"github.com/go-chi/chi/v5"
)

type ReadingHandler struct {
	DB *sql.DB
}

func (h *ReadingHandler) Create(w http.ResponseWriter, r *http.Request) {
	batchID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid batch id", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	reading := models.Reading{
		BatchID:    batchID,
		RecordedAt: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}

	if v := r.FormValue("gravity"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			http.Error(w, "invalid gravity value", http.StatusBadRequest)
			return
		}
		reading.Gravity = sql.NullFloat64{Float64: f, Valid: true}
	}

	if v := r.FormValue("temp_c"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			http.Error(w, "invalid temperature value", http.StatusBadRequest)
			return
		}
		reading.TempC = sql.NullFloat64{Float64: f, Valid: true}
	}

	if v := r.FormValue("taste_notes"); v != "" {
		reading.TasteNotes = sql.NullString{String: v, Valid: true}
	}

	id, err := models.CreateReading(h.DB, reading)
	if err != nil {
		http.Error(w, "failed to create reading", http.StatusInternalServerError)
		return
	}
	reading.ID = int(id)

	if isHTMX(r) {
		tmpl, err := template.ParseFiles("templates/partials/reading_row.html")
		if err != nil {
			http.Error(w, "failed to parse template", http.StatusInternalServerError)
			return
		}
		tmpl.ExecuteTemplate(w, "reading_row", reading)
		return
	}

	http.Redirect(w, r, "/batches/"+strconv.Itoa(batchID), http.StatusSeeOther)
}

func (h *ReadingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid reading id", http.StatusBadRequest)
		return
	}

	if err := models.DeleteReading(h.DB, id); err != nil {
		http.Error(w, "failed to delete reading", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
