package main

import (
	"github.com/DeniskaRediska/weather-bot/db"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

func f1() {}

func main() {

	db.InitDB() //Dont forgot about BD

	bot, err := tgbotapi.NewBotAPI("744595298:AAELmkW03wYJobHCkSePPHWcNqG6gEZVlSc")
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	markup := tgbotapi.InlineKeyboardMarkup{}

	reply := "Привет, я Погодный бот"

	callBackWeatherMenu := "WeatherMenu"
	callBackWeatherRightNow := "WeatherRightNow"
	callBackWeatherNotification := "WeatherNotification"
	callBackWeatherLocation := "WeatherLocation"
	callBackWeatherNotifPush := "WeatherPuch"
	callBackWeatherTime := "WeatherTime"

	markupMenu := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Погода в данный момент", CallbackData: &callBackWeatherRightNow}},
			{{Text: "Настройка уведомлений", CallbackData: &callBackWeatherNotification}},
			{{Text: "Задать локацию", CallbackData: &callBackWeatherLocation}},
		},
	}

	var userStruct db.SettingBot
	var pointUserStruct *db.SettingBot
	pointUserStruct = &userStruct

	var awaitLocation = make(map[int]bool)
	flagStart := false

	for update := range updates {

		chatID := int64(0)

		//Добавления id пользователя в БД
		if update.Message != nil { //"заглушка" от callback'ов
			userStruct.ID = update.Message.From.ID
			if flagStart != true {
				pointUserStruct, err = db.GetSettingById(update.Message.From.ID)
				if err != nil {
					log.Println(err, "Катастрофа!")
				}
				if pointUserStruct == nil {
					_, err := db.InsertSetting(userStruct)
					if err != nil {
						panic(err)
					}
				}
				flagStart = true
			}
		}

		if update.CallbackQuery != nil {
			switch update.CallbackQuery.Data {

			case callBackWeatherMenu:
				markup = markupMenu
				chatID = update.CallbackQuery.Message.Chat.ID

			case callBackWeatherRightNow:
				reply = "Погода в данный момент\nВозврат в меню"
				update.CallbackQuery.Data = callBackWeatherMenu
				chatID = update.CallbackQuery.Message.Chat.ID

			case callBackWeatherNotification:
				reply = "Настройка уведомлений"
				markup = tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
						{{Text: "Получать уведомления", CallbackData: &callBackWeatherNotifPush}},
						{{Text: "Настроить время", CallbackData: &callBackWeatherTime}},
					},
				}
				chatID = update.CallbackQuery.Message.Chat.ID

				//Обрабатываем кнопку "задать локацию"
			case callBackWeatherLocation:
				reply = "Пожалуйста, отправьте мне локацию\n\"Прикрепить\"->\"Location\""
				markup = tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
						{{Text: "Отмена", CallbackData: &callBackWeatherMenu}},
					},
				}
				awaitLocation[update.CallbackQuery.From.ID] = true
				//Ждем от пользователя нужного сообщения
				update.CallbackQuery.Data = callBackWeatherLocation
				chatID = update.CallbackQuery.Message.Chat.ID

			case callBackWeatherNotifPush:
				reply = "Настройка уведомления\nВозврат в меню"
				markup = markupMenu
				chatID = update.CallbackQuery.Message.Chat.ID

			case callBackWeatherTime:
				reply = "Выбор времени\nВозврат в меню"
				markup = markupMenu
				update.CallbackQuery.Data = callBackWeatherMenu
				chatID = update.CallbackQuery.Message.Chat.ID

			}

		} else if update.Message != nil {
			if update.Message.Text == "/start" {
				markup = tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
						{{Text: "Погода", CallbackData: &callBackWeatherMenu}},
					},
				}
				chatID = update.Message.Chat.ID
			}
			if awaitLocation[update.Message.From.ID] {
				if update.Message.Location != nil {

					userStruct.Lat = update.Message.Location.Latitude
					userStruct.Lon = update.Message.Location.Longitude

					_, err = db.UpdateSetting(userStruct)
					if err != nil {
						panic(err)
					}
					reply = "Локация успешно задана\n"
					markup = markupMenu
					awaitLocation[update.Message.From.ID] = false
					chatID = update.Message.Chat.ID
				} else {
					reply = "Пожалуйста, отправьте мне локацию\n\"Прикрепить\"->\"Location\""
					chatID = update.Message.Chat.ID
				}

			}
		}

		if chatID != 0 {
			msg := tgbotapi.NewMessage(chatID, reply)
			msg.ReplyMarkup = &markup
			bot.Send(msg)
		}

	}

}
