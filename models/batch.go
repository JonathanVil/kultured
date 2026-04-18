package models

import (
    "database/sql"
)

type Batch struct {
    ID            int
    Name          string
    StartedAt     string
    TeaType       string
    SugarG        float64
    VolumeL       float64
    ScobyWeightG  float64
    Stage         string
    Notes         sql.NullString
    CreatedAt     string
}

func GetAllBatches(db *sql.DB) ([]Batch, error) {
    rows, err := db.Query(`
        SELECT id, name, started_at, tea_type, sugar_g, volume_l, scoby_weight_g, stage, notes, created_at
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
            &b.ID, &b.Name, &b.StartedAt, &b.TeaType,
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
