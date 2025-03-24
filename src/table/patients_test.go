package table

import (
	_ "github.com/mattn/go-sqlite3"
)

// ** テスト用のデータベースセットアップ **
// func setupTestDB(t *testing.T) *PatientsTable {
// 	t.Helper()

// 	// テスト用のSQLiteデータベースをメモリ上に作成
// 	db, err := sql.Open("sqlite3", ":memory:")
// 	if err != nil {
// 		t.Fatalf("テスト用DBのオープンに失敗: %v", err)
// 	}

// 	// テーブル作成
// 	createTableSQL := `
// 	CREATE TABLE patients (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		hashed_id TEXT,
// 		file_name TEXT,
// 		recorded_date TEXT,
// 		patient_id TEXT,
// 		name TEXT,
// 		number TEXT,
// 		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
// 		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
// 	);
// 	`
// 	_, err = db.Exec(createTableSQL)
// 	if err != nil {
// 		t.Fatalf("テーブル作成失敗: %v", err)
// 	}

// 	return &PatientsTable{DB: db}
// }
