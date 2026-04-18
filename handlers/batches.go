package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/JonathanVil/kultured/models"
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
