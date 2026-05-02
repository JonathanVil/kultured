package models

import "database/sql"

type Note struct {
	ID        int
	BatchID   int
	Note      string
	CreatedAt string
}

func GetNotesForBatch(db *sql.DB, batchID int) ([]Note, error) {
	rows, err := db.Query(`
		SELECT id, batch_id, note, created_at
		FROM batch_notes
		WHERE batch_id = ?
		ORDER BY created_at DESC
	`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.BatchID, &n.Note, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func CreateNote(db *sql.DB, n Note) (int64, error) {
	result, err := db.Exec(
		`INSERT INTO batch_notes (batch_id, note) VALUES (?, ?)`,
		n.BatchID, n.Note,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func DeleteNote(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM batch_notes WHERE id = ?`, id)
	return err
}
