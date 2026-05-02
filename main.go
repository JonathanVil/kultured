package main

import (
	"log"
	"net/http"

	"github.com/JonathanVil/kultured/db"
	"github.com/JonathanVil/kultured/handlers"
	"github.com/go-chi/chi/v5"
)

func main() {
	database, err := db.Open("brew.db")
	if err != nil {
		log.Fatal("could not open database:", err)
	}
	defer database.Close()

	batchHandler := &handlers.BatchHandler{DB: database}
	noteHandler := &handlers.NoteHandler{DB: database}

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Get("/batches", batchHandler.List)
		r.Post("/batches", batchHandler.Create)
		r.Get("/batches/{id}", batchHandler.Get)
		r.Post("/batches/{id}/stage", batchHandler.UpdateStage)
		r.Delete("/batches/{id}", batchHandler.Delete)
		r.Post("/batches/{id}/notes", noteHandler.Create)
		r.Delete("/notes/{id}", noteHandler.Delete)
	})

	log.Println("kultured running on http://localhost:8085")
	log.Fatal(http.ListenAndServe(":8085", r))
}
