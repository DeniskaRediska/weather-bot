package db

import (
	"database/sql"
	_ "database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
)

func init() {

}

const dataSource = ""

type SettingBot struct {
	ID           int
	Lat          float64
	Lon          float64
	Notification bool
	Time         string
}

func GetSettingById(userID int) (*SettingBot, error) {
	getInfo := SettingBot{}
	err := db.Get(&getInfo, "SELECT * FROM setting WHERE id=?", userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &getInfo, nil
}

func InsertSetting(setting SettingBot) (*SettingBot, error) {
	_, err := db.Exec(
		"INSERT INTO setting (id, lat, lon, notification, time) VALUES (?, ?, ?, ?, ?)",
		setting.ID,
		setting.Lat,
		setting.Lon,
		setting.Notification,
		setting.Time,
	)
	return &setting, err
}

func UpdateSetting(setting SettingBot) (*SettingBot, error) {
	_, err := db.Exec(
		"UPDATE setting SET lat=?, lon=?, notification=?, time=? WHERE id=?",
		setting.Lat,
		setting.Lon,
		setting.Notification,
		setting.Time,
		setting.ID,
	)
	return &setting, err
}

var db *sqlx.DB

func InitDB() {
	newDb, err := sqlx.Open("mysql", dataSource)
	if err != nil {
		log.Println("Didnt open data base", err)
		panic(err)
	}
	testReturn, err := newDb.Exec("SELECT 1")
	if err != nil {
		panic(err)
	}

	db = newDb
	log.Println(testReturn, "Database was authorized")
}
