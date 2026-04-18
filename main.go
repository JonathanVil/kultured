package main

import (
    "log"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/JonathanVil/kultured/db"
    "github.com/JonathanVil/kultured/handlers"
)

func main() {
    database, err := db.Open("brew.db")
    if err != nil {
        log.Fatal("could not open database:", err)
    }
    defer database.Close()

    batchHandler := &handlers.BatchHandler{DB: database}
    readingHandler := &handlers.ReadingHandler{DB: database}

    r := chi.NewRouter()
    r.Get("/", batchHandler.Index)
    r.Get("/batches/new", batchHandler.New)
    r.Post("/batches", batchHandler.Create)
    r.Get("/batches/{id}", batchHandler.Show)
    r.Post("/batches/{id}/stage", batchHandler.UpdateStage)
    r.Delete("/batches/{id}", batchHandler.Delete)
    r.Post("/batches/{id}/readings", readingHandler.Create)
    r.Delete("/readings/{id}", readingHandler.Delete)

    log.Println("kultured running on http://localhost:8085")
    log.Fatal(http.ListenAndServe(":8085", r))
}
