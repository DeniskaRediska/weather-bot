package db

import (
	"log"
	"testing"
)

func init() {
	InitDB()
}

func TestCrud(t *testing.T) {
	TestDb := SettingBot{
		ID:           2,
		Lat:          53.1241,
		Lon:          51.6125,
		Notification: true,
		Time:         "9:00",
	}

	_, err := InsertSetting(TestDb)
	if err != nil {
		panic(err)
	}

	testDbCopy := TestDb
	testDbCopy.Notification = false

	_, err = UpdateSetting(testDbCopy)

	bot, err := GetSettingById(1)
	if err != nil {
		panic(err)
	}
	log.Println(*bot)
}
