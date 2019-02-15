package schedule

import (
	"errors"
	"fmt"
	"github.com/DeniskaRediska/weather-bot/cronbot"
	"github.com/DeniskaRediska/weather-bot/db"
	"github.com/robfig/cron"
	"log"
	"strings"
)

var arrayTasks = make(map[int]*cron.Cron)

func LoadNotificationsTasks() {
	settings, err := db.GetEnableSettingsDB()
	if err != nil {
		log.Println(err, "a-a-a")
	}
	for _, setting := range settings {
		cron, err := createTasks(setting.ID, setting.Time)
		if err != nil {
			log.Println("Шо-то с базой", err)
			continue
		}
		arrayTasks[setting.ID] = cron
	}
}

func CreateNotificationsTasks(idUser int, time string) {

	oldCron, ok := arrayTasks[idUser]
	if ok {
		oldCron.Stop()
	}

	cron, err := createTasks(idUser, time)
	if err != nil {
		log.Println("Шо-то с базой", err)
		return
	}
	arrayTasks[idUser] = cron
}

func DeleteNotificationsTask(idUser int) {
	oldCron, ok := arrayTasks[idUser]
	if ok {
		oldCron.Stop()
		delete(arrayTasks, idUser)
	}
}

func splitTime(time string) (string, string, error) {
	timeArray := strings.Split(time, ":")
	if len(timeArray) != 2 {
		return "", "", errors.New("invalide")
	}
	return timeArray[0], timeArray[1], nil
}

func createTasks(idUser int, time string) (*cron.Cron, error) {
	cron := cron.New()
	Hour, Minute, err := splitTime(time)
	if err != nil {
		return nil, err
	}

	err = cron.AddFunc(fmt.Sprintf(
		"0 %s %s * * *", Minute, Hour),
		func() { cronbot.AutoPush(idUser) },
	)
	if err != nil {
		return nil, err
	}
	cron.Start()
	return cron, nil
}
