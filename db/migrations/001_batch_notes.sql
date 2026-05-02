ALTER TABLE batches RENAME COLUMN scoby_weight_g TO scoby_volume_ml;
ALTER TABLE batches RENAME COLUMN volume_l TO tea_volume_l;
ALTER TABLE batches DROP COLUMN notes;
ALTER TABLE batches ADD COLUMN start_f2 TEXT;
ALTER TABLE batches ADD COLUMN done_at TEXT;
DROP TABLE IF EXISTS readings;
CREATE TABLE batch_notes (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    batch_id   INTEGER NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    note       TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
)
