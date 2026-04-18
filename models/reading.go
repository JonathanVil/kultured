package models

import "database/sql"

type Reading struct {
	ID         int
	BatchID    int
	RecordedAt string
	Gravity    sql.NullFloat64
	TempC      sql.NullFloat64
	TasteNotes sql.NullString
}

func GetReadingsForBatch(db *sql.DB, batchID int) ([]Reading, error) {
	rows, err := db.Query(`
		SELECT id, batch_id, recorded_at, gravity, temp_c, taste_notes
		FROM readings
		WHERE batch_id = ?
		ORDER BY recorded_at DESC
	`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []Reading
	for rows.Next() {
		var r Reading
		if err := rows.Scan(&r.ID, &r.BatchID, &r.RecordedAt, &r.Gravity, &r.TempC, &r.TasteNotes); err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	return readings, rows.Err()
}

func GetLatestGravityForBatch(db *sql.DB, batchID int) (float64, bool) {
	var g float64
	err := db.QueryRow(`
		SELECT gravity FROM readings
		WHERE batch_id = ? AND gravity IS NOT NULL
		ORDER BY recorded_at DESC LIMIT 1
	`, batchID).Scan(&g)
	return g, err == nil
}

func CreateReading(db *sql.DB, r Reading) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO readings (batch_id, recorded_at, gravity, temp_c, taste_notes)
		VALUES (?, ?, ?, ?, ?)
	`, r.BatchID, r.RecordedAt, r.Gravity, r.TempC, r.TasteNotes)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func DeleteReading(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM readings WHERE id = ?`, id)
	return err
}
