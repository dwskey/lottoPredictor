// db/lotto_results.go
package db

import (
	"database/sql"
	"log"

	"lottopredictor/internal/fetcher"
)

func CreateLottoResultsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS lotto_results (
			draw_number INTEGER PRIMARY KEY,
			draw_date TEXT,
			n1 INTEGER,
			n2 INTEGER,
			n3 INTEGER,
			n4 INTEGER,
			n5 INTEGER,
			n6 INTEGER,
			bonus INTEGER
		)`)
	return err
}

func SaveDrawResult(db *sql.DB, data *fetcher.DrawData) {
	_, err := db.Exec(`
		INSERT OR IGNORE INTO lotto_results(
			draw_number, draw_date, n1, n2, n3, n4, n5, n6, bonus
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		data.DrwNo, data.DrwNoDate,
		data.DrwtNo1, data.DrwtNo2, data.DrwtNo3, data.DrwtNo4, data.DrwtNo5, data.DrwtNo6,
		data.BnusNo)
	if err != nil {
		log.Printf("[DB] 저장 실패 (회차 %d): %v\n", data.DrwNo, err)
	} else {
		if data.DrwNo%100 == 0 {
			log.Printf("[DB] insering(drawnumber: ~%d)\n", data.DrwNo)
		}
	}
}

func GetLatestDrawNumber(db *sql.DB) int {
	row := db.QueryRow("SELECT MAX(draw_number) FROM lotto_results")
	var max int
	row.Scan(&max)
	return max
}
