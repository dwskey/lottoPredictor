// db/init.go
package db

import (
	"database/sql"
	"errors"

	_ "modernc.org/sqlite"
)

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// 모든 테이블 생성
	if err := CreatePredictionMetaTable(db); err != nil {
		return nil, errors.New("CreatePredictionMetaTable 실패: " + err.Error())
	}

	if err := CreateLottoResultsTable(db); err != nil {
		return nil, errors.New("CreateLottoResultsTable 실패: " + err.Error())
	}

	if err := CreateDrawProbabilitiesTable(db); err != nil {
		return nil, errors.New("CreateDrawProbabilitiesTable 실패: " + err.Error())
	}

	if err := CreateReappearanceProbabilitiesTable(db); err != nil {
		return nil, errors.New("CreateReappearanceProbabilitiesTable 실패: " + err.Error())
	}

	if err := CreatePredictionResultsTable(db); err != nil {
		return nil, errors.New("CreatePredictionResultsTable 실패: " + err.Error())
	}

	return db, nil
}
