# kultured 🫙

A personal kombucha batch tracker. Tracks fermentation stages, brew parameters, and notes per batch.

## Stack

- **Backend** — Go, [chi](https://github.com/go-chi/chi), SQLite (`modernc.org/sqlite`)
- **Frontend** — Svelte, Vite, Tailwind CSS, shadcn-svelte, svelte-spa-router
- **Database** — `brew.db` (SQLite file, created on first run)

## Dev setup

**Prerequisites:** Go, Node

**Backend** (runs on `localhost:8085`):
```sh
go run .
```

**Frontend** (runs on `localhost:5173`, proxies `/api/*` to the backend):
```sh
cd web
npm install
npm run dev
```

Open `http://localhost:5173` during development. The Vite dev server provides hot module reloading and forwards API calls to the Go server automatically.

## Production build

```sh
cd web && npm run build
cd .. && go run .
```

`npm run build` writes the compiled frontend into `web/dist/`, which the Go binary embeds at compile time and serves from `/`. The backend then handles both the API and the static frontend with no separate Node process needed.

## Project structure

```
kultured/
├── main.go                 # Server entry point, routes
├── brew.db                 # SQLite database (gitignored in prod)
├── calc/calc.go            # Fermentation day calculations
├── db/
│   ├── db.go               # DB init and migration runner
│   └── migrations/         # Numbered .sql migration files
├── handlers/
│   ├── batches.go          # GET/POST/DELETE /api/batches[/{id}]
│   └── notes.go            # POST /api/batches/{id}/notes, DELETE /api/notes/{id}
├── models/
│   ├── batch.go            # Batch struct and queries
│   └── note.go             # Note struct and queries
└── web/                    # Svelte frontend
    ├── src/
    │   ├── pages/          # BatchList, BatchDetail, NewBatch
    │   ├── lib/components/ # shadcn-svelte components
    │   └── routes.js       # svelte-spa-router route map
    └── dist/               # Built frontend (embedded by Go)
```

## Adding a migration

1. Create `db/migrations/<NNN>_description.sql` where `NNN` follows the last number in sequence.
2. Write standard SQLite DDL — each statement separated by `;`.
3. Restart the server. Migrations are applied automatically at startup and tracked in the `schema_migrations` table.

Migrations run in filename order and are skipped if already applied, so they are safe to run on an existing database.
