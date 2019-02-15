package cronbot

import (
	"fmt"
	"github.com/DeniskaRediska/weather-bot/db"
	"github.com/DeniskaRediska/weather-bot/openweather"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"regexp"
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
		log.Fatal(err)
	}

	temp, err := openweather.WeatherScheule(userInfo.Lat, userInfo.Lon)
	if err != nil {
		log.Println(err)
		return
	}

	try, err := regexp.Compile("\\d{2}:\\d{2}")
	if err != nil {
		log.Println(err)
	}
	msg := fmt.Sprintln(
		"Погода на день для города ", temp.City.Name, ":",
		"\n В ", try.FindString(temp.List[0].Dt_txt), " будет ", temp.List[0].Main.Temp, " градусов",
		"\n В ", try.FindString(temp.List[1].Dt_txt), " будет ", temp.List[1].Main.Temp, " градусов",
		"\n В ", try.FindString(temp.List[2].Dt_txt), " будет ", temp.List[2].Main.Temp, " градусов",
		"\n В ", try.FindString(temp.List[3].Dt_txt), " будет ", temp.List[3].Main.Temp, " градусов",
		"\n В ", try.FindString(temp.List[4].Dt_txt), " будет ", temp.List[4].Main.Temp, " градусов",
		"\n В ", try.FindString(temp.List[5].Dt_txt), " будет ", temp.List[5].Main.Temp, " градусов",
		"\n В ", try.FindString(temp.List[6].Dt_txt), " будет ", temp.List[6].Main.Temp, " градусов",
		"\n В ", try.FindString(temp.List[7].Dt_txt), " будет ", temp.List[7].Main.Temp, " градусов",
		"\nЕсли хотите отключить уведомления или изменить локацию, зайдите в меню через команду /start",
	)
	message := tgbotapi.NewMessage(int64(userInfo.ID), msg)
	bot.Send(message)
}
