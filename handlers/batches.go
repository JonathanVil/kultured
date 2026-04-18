package handlers

import (
    "database/sql"
    "encoding/json"
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

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(batches)
}
