package db

import (
	"database/sql"
)

func CreatePredictionMetaTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS prediction_meta (
			draw_number INTEGER,
			idx INTEGER,
			created_at TEXT,
			PRIMARY KEY (draw_number, idx)
		)`)
	return err
}

func InsertPredictionMeta(db *sql.DB, drawNo int) (int, error) {
	var currentMax sql.NullInt64
	row := db.QueryRow("SELECT MAX(idx) FROM prediction_meta WHERE draw_number = ?", drawNo)
	err := row.Scan(&currentMax)
	if err != nil {
		return 0, err
	}

	newIdx := 1
	if currentMax.Valid {
		newIdx = int(currentMax.Int64) + 1
	}

	stmt, err := db.Prepare(`
		INSERT INTO prediction_meta(draw_number, idx, created_at)
		VALUES (?, ?, datetime('now'))
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(drawNo, newIdx)
	if err != nil {
		return 0, err
	}
	return newIdx, nil
}
