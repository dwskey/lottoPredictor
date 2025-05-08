// db/prediction_results.go
package db

import (
	"database/sql"
	"fmt"
	"lottopredictor/internal/common"
)

func CreatePredictionResultsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS prediction_results (
			draw_number INTEGER,
			meta_idx INTEGER,
			set_index INTEGER,
			num1 INTEGER,
			num2 INTEGER,
			num3 INTEGER,
			num4 INTEGER,
			num5 INTEGER,
			num6 INTEGER,
			percentage REAL,
			rank INTEGER,
			created_at TEXT,
			PRIMARY KEY (draw_number, meta_idx, set_index)
		)`)
	return err
}

func SavePredictionResults(db *sql.DB, metaID int64, drawNo int, predictions [][]int) error {
	stmt, err := db.Prepare(`
		INSERT INTO prediction_results
		(draw_number, meta_idx, set_index, num1, num2, num3, num4, num5, num6, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, set := range predictions {
		if len(set) != 6 {
			return fmt.Errorf("invalid set length: %v", set)
		}
		_, err := stmt.Exec(metaID, drawNo, i+1, set[0], set[1], set[2], set[3], set[4], set[5])
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdatePredictionEvaluations(db *sql.DB, drawNo int, actual []int, bonus int) error {
	set := make(map[int]bool)
	for _, n := range actual {
		set[n] = true
	}

	query := `
		SELECT meta_idx, set_index, num1, num2, num3, num4, num5, num6
		FROM prediction_results
		WHERE draw_number = ?
	`
	rows, err := db.Query(query, drawNo-1)
	if err != nil {
		return err
	}
	defer rows.Close()

	updateStmt, err := db.Prepare(`
		UPDATE prediction_results
		SET percentage = ?, rank = ?
		WHERE draw_number = ? AND meta_idx = ? AND set_index = ?
	`)
	if err != nil {
		return err
	}
	defer updateStmt.Close()

	for rows.Next() {
		var metaIdx, setIdx, n1, n2, n3, n4, n5, n6 int
		rows.Scan(&metaIdx, &setIdx, &n1, &n2, &n3, &n4, &n5, &n6)

		nums := []int{n1, n2, n3, n4, n5, n6}
		matched := 0
		bonusMatched := false

		for _, n := range nums {
			if set[n] {
				matched++
			}
			if n == bonus {
				bonusMatched = true
			}
		}

		// 퍼센트 계산
		percent := float64(matched) / 6.0 * 100
		rank := common.RankNone
		switch matched {
		case 6:
			rank = common.RankFirst
		case 5:
			if bonusMatched {
				rank = common.RankSecond
			} else {
				rank = common.RankThird
			}
		case 4:
			rank = common.RankFourth
		case 3:
			rank = common.RankFifth
		default:
			rank = common.RankNone
		}

		updateStmt.Exec(percent, rank, drawNo-1, metaIdx, setIdx)
	}

	return nil
}
