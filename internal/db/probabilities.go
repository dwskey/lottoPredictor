// db/probabilities.go
package db

import (
	"database/sql"
	"fmt"
)

func CreateDrawProbabilitiesTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS draw_probabilities (
			draw_number INTEGER,
			number INTEGER,
			probability REAL
		)`)
	return err
}

func CreateReappearanceProbabilitiesTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS reappearance_probabilities (
			draw_number INTEGER,
			number INTEGER,
			probability REAL
		)`)
	return err
}

func SaveDrawProbabilities(db *sql.DB, drawNo int, probs map[int]float64) {
	row := db.QueryRow("SELECT COUNT(1) FROM draw_probabilities WHERE draw_number = ?", drawNo)
	var exists int
	row.Scan(&exists)

	stmt, _ := db.Prepare("INSERT INTO draw_probabilities(draw_number, number, probability) VALUES (?, ?, ?)")
	defer stmt.Close()
	for num, prob := range probs {
		if exists == 0 {
			stmt.Exec(drawNo, num, fmt.Sprintf("%.3f", prob))
		}
	}
}

func SaveReappearanceProbabilities(db *sql.DB, drawNo int, probs map[int]float64) {
	stmt, _ := db.Prepare("INSERT INTO reappearance_probabilities(draw_number, number, probability) VALUES (?, ?, ?)")
	defer stmt.Close()
	for num, prob := range probs {
		stmt.Exec(drawNo, num, fmt.Sprintf("%.3f", prob))
	}
}
