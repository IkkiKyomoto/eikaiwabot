package database

//DBに対する操作を記述

import (
	"database/sql"
)

// 取り出すROWの件数
const (
	numberOfActs = 11
)

// DBで扱うActivity構造体を定義
type Activity struct {
	UserID  string
	Role    string
	Message string
}

// アクティビティを追加
func InsertRow(db *sql.DB, act Activity) error {
	_, err := db.Exec("INSERT INTO Activity (message, role, userid) VALUES ($1, $2, $3)", act.Message, act.Role, act.UserID)
	return err
}

// 過去のアクティビティを取得
func GetRows(db *sql.DB, userID string) ([]Activity, error) {
	rows, err := db.Query("SELECT message, role FROM Activity WHERE userid = $1 ORDER BY id LIMIT $2", userID, numberOfActs)
	if err != nil {
		return []Activity{}, err
	}
	defer rows.Close()

	activities := []Activity{}
	for rows.Next() {
		a := Activity{UserID: userID}
		err := rows.Scan(&a.Message, &a.Role)
		if err != nil {
			return []Activity{}, err
		}
		activities = append(activities, a)
	}
	return activities, err
}
