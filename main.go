package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/JonathanVil/kultured/db"
	"github.com/JonathanVil/kultured/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

//go:embed all:web/dist
var webDist embed.FS

func main() {
	database, err := db.Open("brew.db")
	if err != nil {
		log.Fatal("could not open database:", err)
	}
	defer database.Close()

	batchHandler := &handlers.BatchHandler{DB: database}
	noteHandler := &handlers.NoteHandler{DB: database}

	r := chi.NewRouter()

	r.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}).Handler)

	r.Route("/api", func(r chi.Router) {
		r.Get("/batches", batchHandler.List)
		r.Post("/batches", batchHandler.Create)
		r.Get("/batches/{id}", batchHandler.Get)
		r.Put("/batches/{id}", batchHandler.Update)
		r.Post("/batches/{id}/stage", batchHandler.UpdateStage)
		r.Delete("/batches/{id}", batchHandler.Delete)
		r.Post("/batches/{id}/notes", noteHandler.Create)
		r.Delete("/notes/{id}", noteHandler.Delete)
	})

	distFS, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		log.Fatal("could not sub web/dist:", err)
	}
	r.Handle("/*", spaHandler(distFS))

	log.Println("kultured running on http://localhost:8085")
	log.Fatal(http.ListenAndServe(":8085", r))
}

// spaHandler serves static files from fsys and falls back to index.html for
// any path that doesn't match a real file (supporting client-side routing).
func spaHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(fsys))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fsys.Open(path); err != nil {
			// Unknown path — let the SPA router handle it client-side.
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
