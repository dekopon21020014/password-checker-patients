package table

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// **PatientsTable 構造体**
type PatientsTable struct {
	DB *sql.DB
}

// **データベースのセットアップ**
func GetDatabase(dsn string) *PatientsTable {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalf("データベースのオープンに失敗: %v", err)
	}
	return &PatientsTable{DB: db}
}

func (p *PatientsTable) CheckPassword(password string) (map[string]map[string]int, error) {
	result := make(map[string]map[string]int)

	rows, err := p.DB.Query("SELECT patient_id, hashed_id, created_at FROM patients")
	if err != nil {
		log.Println("Error querying database:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var patientID, hashedID, createdAt string
		if err := rows.Scan(&patientID, &hashedID, &createdAt); err != nil {
			log.Println("Error scanning row:", err)
			continue
		}

		// 日付をyyyy/mm形式に変換
		parsedTime, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			log.Println("Error parsing date:", err)
			continue
		}
		dateKey := parsedTime.Format("2006/01/02")

		hashedInput := sha256Hash(patientID, password)
		matchType := "一致"
		if hashedInput != hashedID {
			matchType = "不一致"
		}

		if _, exists := result[dateKey]; !exists {
			result[dateKey] = map[string]int{"一致": 0, "不一致": 0}
		}
		result[dateKey][matchType]++
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
		return nil, err
	}

	return result, nil
}

// ** SHA256 ハッシュ関数 **
func sha256Hash(patientID, password string) string {
	hash := sha256.Sum256([]byte(patientID + password))
	return hex.EncodeToString(hash[:])
}
