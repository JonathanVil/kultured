package models

import (
    "database/sql"
)

type Batch struct {
    ID            int
    Name          string
    StartedAt     string
    TeaType       string
    TeaG          float64
    SteepMin      float64
    SugarG        float64
    VolumeL       float64
    ScobyWeightG  float64
    Stage         string
    Notes         sql.NullString
    CreatedAt     string
}

func GetAllBatches(db *sql.DB) ([]Batch, error) {
    rows, err := db.Query(`
        SELECT id, name, started_at, tea_type, tea_g, steep_min, sugar_g, volume_l, scoby_weight_g, stage, notes, created_at
        FROM batches
        ORDER BY started_at DESC
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var batches []Batch
    for rows.Next() {
        var b Batch
        err := rows.Scan(
            &b.ID, &b.Name, &b.StartedAt, &b.TeaType, &b.TeaG, &b.SteepMin,
            &b.SugarG, &b.VolumeL, &b.ScobyWeightG,
            &b.Stage, &b.Notes, &b.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        batches = append(batches, b)
    }
    return batches, rows.Err()
}

func CreateBatch(db *sql.DB, b Batch) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO batches
		(name, started_at, tea_type, tea_g, steep_min, sugar_g, volume_l, scoby_weight_g, stage, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, b.Name, b.StartedAt, b.TeaType, b.TeaG, b.SteepMin, b.SugarG, b.VolumeL, b.ScobyWeightG, b.Stage, b.Notes)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func GetBatch(db *sql.DB, id int) (Batch, error) {
	var b Batch
	err := db.QueryRow(`
		SELECT id, name, started_at, tea_type, tea_g, steep_min, sugar_g, volume_l, scoby_weight_g, stage, notes, created_at
		FROM batches WHERE id = ?
	`, id).Scan(&b.ID, &b.Name, &b.StartedAt, &b.TeaType, &b.TeaG, &b.SteepMin, &b.SugarG, &b.VolumeL, &b.ScobyWeightG, &b.Stage, &b.Notes, &b.CreatedAt)
	return b, err
}

func UpdateStage(db *sql.DB, id int, stage string) error {
	_, err := db.Exec(`UPDATE batches SET stage = ? WHERE id = ?`, stage, id)
	return err
}

func DeleteBatch(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM batches WHERE id = ?`, id)
	return err
}
