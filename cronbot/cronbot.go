package cronbot

import (
	"fmt"
	"github.com/DeniskaRediska/weather-bot/db"
	"github.com/DeniskaRediska/weather-bot/openweather"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"regexp"
	"strings"
)

const TGAPI = "744595298:AAELmkW03wYJobHCkSePPHWcNqG6gEZVlSc"

func AutoPush(idUser int) {

	userInfo, err := db.GetSettingById(idUser)
	if err != nil {
		log.Println(err, "\nНе удалось получить информацию из базы данных")
		return
	}

	bot, err := tgbotapi.NewBotAPI(TGAPI)
	if err != nil {
		log.Println(err)
		return
	}

	temp, err := openweather.GetDailyForecast(userInfo.Lat, userInfo.Lon)
	if err != nil {
		log.Println(err)
		return
	}

	try, err := regexp.Compile("\\d{2}:\\d{2}")
	if err != nil {
		log.Println(err)
	}

	array := make([]string, len(temp.List))
	for i := range temp.List {
		array[i] = fmt.Sprintln("В ", try.FindString(temp.List[i].Dt_txt), " будет ", temp.List[i].Main.Temp, " градусов")
	}

	msg := fmt.Sprintln(
		"Погода на день для города ", temp.City.Name, ":\n",
		strings.Join(array, "\n"),
		"\nЕсли хотите отключить уведомления или изменить локацию, зайдите в меню через команду /start",
	)
	message := tgbotapi.NewMessage(int64(userInfo.ID), msg)
	_, err = bot.Send(message)
	if err != nil {
		return
	}
}
