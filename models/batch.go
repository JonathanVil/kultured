package models

import (
	"database/sql"
	"time"
)

type Batch struct {
	ID                 int
	Name               string
	StartedAt          string
	TeaType            string
	TeaG               float64
	SteepMin           float64
	SugarG             float64
	TeaVolumeMl        float64
	ScobyVolumeMl      float64
	Stage              string
	StartF2            sql.NullString
	DoneAt             sql.NullString
	CreatedAt          string
	ReminderEnabled      bool
	ReminderIntervalDays int
	ReminderTime         string
	ReminderDayOfWeek    sql.NullInt64
	LastRemindedAt       sql.NullString
}

func UpdateBatch(db *sql.DB, b Batch) error {
	// When reverting stage, clear timestamps that belong to later stages.
	// f1 → clear start_f2 and done_at
	// f2 → clear done_at only (start_f2 kept or stays null)
	// bottled/done → no timestamp clearing
	switch b.Stage {
	case "f1":
		_, err := db.Exec(`
			UPDATE batches
			SET name=?, tea_type=?, tea_g=?, steep_min=?, sugar_g=?,
			    tea_volume_ml=?, scoby_volume_ml=?, started_at=?,
			    stage=?, start_f2=NULL, done_at=NULL,
			    reminder_enabled=?, reminder_interval_days=?, reminder_time=?, reminder_day_of_week=?
			WHERE id=?
		`, b.Name, b.TeaType, b.TeaG, b.SteepMin, b.SugarG,
			b.TeaVolumeMl, b.ScobyVolumeMl, b.StartedAt, b.Stage,
			b.ReminderEnabled, b.ReminderIntervalDays, b.ReminderTime, b.ReminderDayOfWeek, b.ID)
		return err
	case "f2":
		_, err := db.Exec(`
			UPDATE batches
			SET name=?, tea_type=?, tea_g=?, steep_min=?, sugar_g=?,
			    tea_volume_ml=?, scoby_volume_ml=?, started_at=?,
			    stage=?, done_at=NULL,
			    reminder_enabled=?, reminder_interval_days=?, reminder_time=?, reminder_day_of_week=?
			WHERE id=?
		`, b.Name, b.TeaType, b.TeaG, b.SteepMin, b.SugarG,
			b.TeaVolumeMl, b.ScobyVolumeMl, b.StartedAt, b.Stage,
			b.ReminderEnabled, b.ReminderIntervalDays, b.ReminderTime, b.ReminderDayOfWeek, b.ID)
		return err
	default:
		_, err := db.Exec(`
			UPDATE batches
			SET name=?, tea_type=?, tea_g=?, steep_min=?, sugar_g=?,
			    tea_volume_ml=?, scoby_volume_ml=?, started_at=?,
			    stage=?,
			    reminder_enabled=?, reminder_interval_days=?, reminder_time=?, reminder_day_of_week=?
			WHERE id=?
		`, b.Name, b.TeaType, b.TeaG, b.SteepMin, b.SugarG,
			b.TeaVolumeMl, b.ScobyVolumeMl, b.StartedAt, b.Stage,
			b.ReminderEnabled, b.ReminderIntervalDays, b.ReminderTime, b.ReminderDayOfWeek, b.ID)
		return err
	}
}

const batchColumns = ` id, name, started_at, tea_type, tea_g, steep_min, sugar_g,
	tea_volume_ml, scoby_volume_ml, stage, start_f2, done_at, created_at,
	reminder_enabled, reminder_interval_days, reminder_time, reminder_day_of_week, last_reminded_at `

func scanBatch(s interface{ Scan(...any) error }) (Batch, error) {
	var b Batch
	err := s.Scan(
		&b.ID, &b.Name, &b.StartedAt, &b.TeaType, &b.TeaG, &b.SteepMin, &b.SugarG,
		&b.TeaVolumeMl, &b.ScobyVolumeMl, &b.Stage, &b.StartF2, &b.DoneAt, &b.CreatedAt,
		&b.ReminderEnabled, &b.ReminderIntervalDays, &b.ReminderTime, &b.ReminderDayOfWeek, &b.LastRemindedAt,
	)
	return b, err
}

func GetAllBatches(db *sql.DB) ([]Batch, error) {
	rows, err := db.Query(`SELECT` + batchColumns + `FROM batches ORDER BY started_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []Batch
	for rows.Next() {
		b, err := scanBatch(rows)
		if err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}
	return batches, rows.Err()
}

func GetBatch(db *sql.DB, id int) (Batch, error) {
	row := db.QueryRow(`SELECT`+batchColumns+`FROM batches WHERE id = ?`, id)
	return scanBatch(row)
}

func GetBatchesWithReminders(db *sql.DB) ([]Batch, error) {
	rows, err := db.Query(`SELECT` + batchColumns + `FROM batches WHERE reminder_enabled = 1 AND stage != 'done'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []Batch
	for rows.Next() {
		b, err := scanBatch(rows)
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
			(name, started_at, tea_type, tea_g, steep_min, sugar_g, tea_volume_ml, scoby_volume_ml, stage)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, b.Name, b.StartedAt, b.TeaType, b.TeaG, b.SteepMin, b.SugarG, b.TeaVolumeMl, b.ScobyVolumeMl, b.Stage)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateStage(db *sql.DB, id int, stage string) error {
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	switch stage {
	case "f2":
		_, err := db.Exec(`UPDATE batches SET stage = ?, start_f2 = ? WHERE id = ?`, stage, now, id)
		return err
	case "done":
		_, err := db.Exec(`UPDATE batches SET stage = ?, done_at = ? WHERE id = ?`, stage, now, id)
		return err
	default:
		_, err := db.Exec(`UPDATE batches SET stage = ? WHERE id = ?`, stage, id)
		return err
	}
}

func UpdateReminderSent(db *sql.DB, id int, ts string) error {
	_, err := db.Exec(`UPDATE batches SET last_reminded_at = ? WHERE id = ?`, ts, id)
	return err
}

func DeleteBatch(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM batches WHERE id = ?`, id)
	return err
}
