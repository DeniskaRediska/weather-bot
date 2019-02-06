package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
)

const TGAPI = "744595298:AAELmkW03wYJobHCkSePPHWcNqG6gEZVlSc"
const OpenWeatherAPI = "23c6fd16f52d2bbd5edc293bfab7681a"

func weatherAnswer(lat float64, lon float64) string {
	client := http.Client{}
	WeatherApi := fmt.Sprint("http://api.openweathermap.org/data/2.5/weather?lat=", lat, "&lon=", lon, "&units=metric&APPID=", OpenWeatherAPI)

	WeatherResponse, err := client.Get(WeatherApi)
	if err != nil {
		//log.Println(err)
		return "Не удалось считать данные"
	}

	DateWeather, err := ioutil.ReadAll(WeatherResponse.Body)
	if err != nil {
		log.Println(err)
	}

	type OWMResponse struct {
		Main struct {
			Temp float64
		}
		Name string
	}
	var WeatherMessageGet OWMResponse

	err = json.Unmarshal(DateWeather, &WeatherMessageGet)
	if err != nil {
		log.Println(err)
	}

	if WeatherMessageGet.Name == "" {
		WeatherMessageGet.Name = "Unknown"
	}

	return fmt.Sprint("Ваш город - ", WeatherMessageGet.Name, "\n Температура ", WeatherMessageGet.Main.Temp)
}

func main() {
	bot, err := tgbotapi.NewBotAPI(TGAPI)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		reply := "Не знаю что сказать"

		if update.Message.Location != nil {
			//log.Println(update.Message.Location)
			reply = weatherAnswer(update.Message.Location.Latitude, update.Message.Location.Longitude)
		}

		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}
