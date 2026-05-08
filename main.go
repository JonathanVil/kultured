package main

import (
	"crypto/subtle"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/JonathanVil/kultured/db"
	"github.com/JonathanVil/kultured/handlers"
	"github.com/JonathanVil/kultured/notify"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

//go:embed all:web/dist
var webDist embed.FS

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "brew.db"
	}
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatal("could not open database:", err)
	}
	defer database.Close()

	ntfyCfg := notify.Config{
		URL:   os.Getenv("NTFY_URL"),
		Topic: os.Getenv("NTFY_TOPIC"),
		User:  os.Getenv("NTFY_USER"),
		Pass:  os.Getenv("NTFY_PASS"),
	}
	if ntfyCfg.Enabled() {
		log.Printf("ntfy reminders enabled: %s/%s", ntfyCfg.URL, ntfyCfg.Topic)
		notify.StartScheduler(database, ntfyCfg)
	} else {
		log.Println("ntfy reminders disabled (NTFY_URL/NTFY_TOPIC not set)")
	}

	batchHandler := &handlers.BatchHandler{DB: database}
	noteHandler := &handlers.NoteHandler{DB: database}
	configHandler := &handlers.ConfigHandler{NtfyEnabled: ntfyCfg.Enabled()}

	r := chi.NewRouter()

	if user, pass := os.Getenv("AUTH_USER"), os.Getenv("AUTH_PASS"); user != "" && pass != "" {
		r.Use(basicAuth(user, pass))
		log.Println("basic auth enabled")
	} else {
		log.Println("warning: AUTH_USER/AUTH_PASS not set — running without authentication")
	}

	r.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}).Handler)

	r.Route("/api", func(r chi.Router) {
		r.Get("/config", configHandler.Get)
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

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8085"
	}
	log.Printf("kultured running on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

// basicAuth is a middleware that enforces HTTP Basic Auth using constant-time
// comparison to prevent timing attacks.
func basicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			userMatch := subtle.ConstantTimeCompare([]byte(u), []byte(username)) == 1
			passMatch := subtle.ConstantTimeCompare([]byte(p), []byte(password)) == 1
			if !ok || !userMatch || !passMatch {
				w.Header().Set("WWW-Authenticate", `Basic realm="kultured", charset="UTF-8"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
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
