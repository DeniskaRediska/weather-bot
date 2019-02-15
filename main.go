package main

import (
	"fmt"
	"github.com/DeniskaRediska/weather-bot/controller"
	"github.com/DeniskaRediska/weather-bot/db"
	"github.com/DeniskaRediska/weather-bot/openweather"
	"github.com/DeniskaRediska/weather-bot/schedule"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"time"
)

const TGAPI = "744595298:AAELmkW03wYJobHCkSePPHWcNqG6gEZVlSc"

var (
	callBackWeatherMenu         = "WeatherMenu"
	callBackWeatherRightNow     = "WeatherRightNow"
	callBackWeatherNotification = "WeatherNotification"
	callBackWeatherLocation     = "WeatherLocation"
	callBackWeatherNotifPush    = "WeatherPuch"
	callBackWeatherNotifTime    = "WeatherTime"

	awaitLocation = make(map[int]bool)
	awaitTime     = make(map[int]bool)
)

func main() {

	db.InitDB()
	schedule.LoadNotificationsTasks()

	bot, err := tgbotapi.NewBotAPI(TGAPI)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	botController := controller.BotController{Bot: bot}
	botController.When(isAwaitLocationYes, handleAwaitYes)
	botController.When(isAwaitLocationNo, handleAwaitNo)
	botController.When(isAwaitTime, handleAwaitTime)
	botController.When(isStart, handleStart)
	botController.When(isLocationButton, handleLocation)
	botController.When(isToMenuButton, backToMenu)
	botController.When(isTempRightNowButton, handleTempRightNow)
	botController.When(isNotificationButton, handleNotificationButton)
	botController.When(isNotifSetting, handleNotifPush)
	botController.When(isTimeSetting, handleTimeSetting)

	for update := range updates {
		botController.HandleUpdate(update)
	}
}

func fineToCreateDB(update tgbotapi.Update) {
	var id int
	if update.Message != nil {
		id = update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		id = update.CallbackQuery.From.ID
	}
	userStruct, err := db.GetSettingById(id)
	if err != nil {
		panic(err)
	}
	if userStruct == nil {
		userStruct = &db.SettingBot{
			ID:           id,
			Lon:          0.0,
			Lat:          0.0,
			Notification: false,
			Time:         "9:00",
		}
		_, err := db.InsertSetting(*userStruct)
		if err != nil {
			panic(err)
		}
	}
}

func isStart(update tgbotapi.Update) bool {
	isStart := update.Message != nil && update.Message.Text == "/start"
	return isStart
}
func handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	fineToCreateDB(update)

	message := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет, я погодный бот")
	message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Погода в данный момент", CallbackData: &callBackWeatherRightNow}},
			{{Text: "Настройка уведомлений", CallbackData: &callBackWeatherNotification}},
			{{Text: "Задать локацию", CallbackData: &callBackWeatherLocation}},
		},
	}
	if _, err := bot.Send(message); err != nil {
		log.Println(err)
	}
}

func isLocationButton(update tgbotapi.Update) bool {
	isLocationButton := update.CallbackQuery != nil && update.CallbackQuery.Data == callBackWeatherLocation
	return isLocationButton
}
func handleLocation(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пожалуйста, отправьте мне свою локацию\n\"Прикрепить\"->\"Location\"")
	message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Отмена", CallbackData: &callBackWeatherMenu}},
		},
	}
	if _, err := bot.Send(message); err != nil {
		log.Println(err)
	}
	awaitLocation[update.CallbackQuery.From.ID] = true
}

func isAwaitLocationYes(update tgbotapi.Update) bool {
	isAwaitLocation := update.Message != nil && awaitLocation[update.Message.From.ID] == true && update.Message.Location != nil
	return isAwaitLocation
}
func handleAwaitYes(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	fineToCreateDB(update)

	dbPointUserS, err := db.GetSettingById(update.Message.From.ID)
	if err != nil {
		panic(err)
	}
	dbPointUserS.Lon = update.Message.Location.Longitude
	dbPointUserS.Lat = update.Message.Location.Latitude
	_, err = db.UpdateSetting(*dbPointUserS)
	if err != nil {
		panic(err)
	}
	message := tgbotapi.NewMessage(update.Message.Chat.ID, "Локация успешно задана")
	message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Погода в данный момент", CallbackData: &callBackWeatherRightNow}},
			{{Text: "Настройка уведомлений", CallbackData: &callBackWeatherNotification}},
			{{Text: "Задать локацию", CallbackData: &callBackWeatherLocation}},
		},
	}
	if _, err := bot.Send(message); err != nil {
		log.Println(err)
	}
	awaitLocation[update.Message.From.ID] = false
}

func isAwaitLocationNo(update tgbotapi.Update) bool {
	isAwaitLocation := update.Message != nil && awaitLocation[update.Message.From.ID] == true && update.Message.Location == nil
	return isAwaitLocation
}
func handleAwaitNo(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	message := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, отправьте мне локацию\n\"Прикрепить\"->\"Location\"")
	message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Отмена", CallbackData: &callBackWeatherMenu}},
		},
	}
	if _, err := bot.Send(message); err != nil {
		log.Println(err)
	}
}

func isToMenuButton(update tgbotapi.Update) bool {
	backToMenu := update.CallbackQuery != nil && update.CallbackQuery.Data == callBackWeatherMenu
	return backToMenu
}
func backToMenu(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	awaitTime[update.CallbackQuery.From.ID] = false
	awaitLocation[update.CallbackQuery.From.ID] = false

	message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Отмена. Возврат в меню")
	message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Погода в данный момент", CallbackData: &callBackWeatherRightNow}},
			{{Text: "Настройка уведомлений", CallbackData: &callBackWeatherNotification}},
			{{Text: "Задать локацию", CallbackData: &callBackWeatherLocation}},
		},
	}
	if _, err := bot.Send(message); err != nil {
		log.Println(err)
	}
}

func isTempRightNowButton(update tgbotapi.Update) bool {
	TempRightNow := update.CallbackQuery != nil && update.CallbackQuery.Data == callBackWeatherRightNow
	return TempRightNow
}
func handleTempRightNow(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	fineToCreateDB(update)

	UserStruct, err := db.GetSettingById(update.CallbackQuery.From.ID)
	if err != nil {
		log.Println(err, "не удалось считать выполнить запрос в базу данных")
		return
	}

	if UserStruct.Lat == 0 && UserStruct.Lon == 0 {
		message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пожалуйста, отправьте мне свою локацию\n\"Прикрепить\"->\"Location\"")
		message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{{Text: "Отмена", CallbackData: &callBackWeatherMenu}},
			},
		}
		if _, err := bot.Send(message); err != nil {
			log.Println(err)
		}
		awaitLocation[update.CallbackQuery.From.ID] = true
		return
	} else {
		cityName, temp, err := openweather.WeatherAnswer(UserStruct.Lat, UserStruct.Lon)
		if err != nil {
			log.Println(err, "Не удалось получить ответ от OpenWeather")
			return
		}
		msg := fmt.Sprintln("Ваш город - ", cityName, "\nТемпература - ", temp)
		message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, msg)
		message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{{Text: "Погода в данный момент", CallbackData: &callBackWeatherRightNow}},
				{{Text: "Настройка уведомлений", CallbackData: &callBackWeatherNotification}},
				{{Text: "Задать локацию", CallbackData: &callBackWeatherLocation}},
			},
		}
		if _, err := bot.Send(message); err != nil {
			log.Println(err)
		}
	}

}

func isNotificationButton(update tgbotapi.Update) bool {
	isNotificationButton := update.CallbackQuery != nil && update.CallbackQuery.Data == callBackWeatherNotification
	return isNotificationButton
}
func handleNotificationButton(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	fineToCreateDB(update)

	userInfo, err := db.GetSettingById(update.CallbackQuery.From.ID)
	if err != nil {
		panic(err)
	}
	msg := "Настройка уведомлений"
	if userInfo.Notification {
		msg = fmt.Sprintln(msg, "\nУведомления включены")
	} else {
		msg = fmt.Sprintln(msg, "\nУведомления отключены")
	}
	msg = fmt.Sprintln(msg, "Текущее время: ", userInfo.Time)
	message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, msg)
	message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Получать уведомления", CallbackData: &callBackWeatherNotifPush}},
			{{Text: "Установить время", CallbackData: &callBackWeatherNotifTime}},
			{{Text: "Отмена", CallbackData: &callBackWeatherMenu}},
		},
	}
	if _, err := bot.Send(message); err != nil {
		log.Println(err)
	}

}

func isNotifSetting(update tgbotapi.Update) bool {
	isNotif := update.CallbackQuery != nil && update.CallbackQuery.Data == callBackWeatherNotifPush
	return isNotif
}
func handleNotifPush(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userInfo, err := db.GetSettingById(update.CallbackQuery.From.ID)
	if err != nil {
		panic(err)
	}
	var msg string
	if userInfo.Notification {
		userInfo.Notification = false
		_, err = db.UpdateSetting(*userInfo)
		if err != nil {
			panic(err)
		}
		schedule.DeleteNotificationsTask(userInfo.ID)
		msg = fmt.Sprintln("Уведомления были отключены")
	} else {
		userInfo.Notification = true
		_, err = db.UpdateSetting(*userInfo)
		if err != nil {
			panic(err)
		}
		schedule.CreateNotificationsTasks(userInfo.ID, userInfo.Time)
		msg = fmt.Sprintln("Уведомления были включены")
	}
	message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, msg)
	message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Получать уведомления", CallbackData: &callBackWeatherNotifPush}},
			{{Text: "Установить время", CallbackData: &callBackWeatherNotifTime}},
			{{Text: "Отмена", CallbackData: &callBackWeatherMenu}},
		},
	}
	if _, err := bot.Send(message); err != nil {
		log.Println(err)
	}
}

func isTimeSetting(update tgbotapi.Update) bool {
	isTimeSetting := update.CallbackQuery != nil && update.CallbackQuery.Data == callBackWeatherNotifTime
	return isTimeSetting
}
func handleTimeSetting(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пожалуйста, отправьте мне время в формате XX:XX")
	message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Отмена", CallbackData: &callBackWeatherMenu}},
		},
	}
	if _, err := bot.Send(message); err != nil {
		log.Println(err)
	}
	awaitTime[update.CallbackQuery.From.ID] = true
}

func isAwaitTime(update tgbotapi.Update) bool {
	isAwaitTime := update.Message != nil && awaitTime[update.Message.From.ID] == true
	return isAwaitTime
}
func handleAwaitTime(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	fineToCreateDB(update)
	userTime, err := time.Parse("15:04", update.Message.Text)
	if err != nil {
		log.Println(err, "не получилось")
		message := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, отправьтьте время в формате\nXX:XX")
		message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{{Text: "Отмена", CallbackData: &callBackWeatherMenu}},
			},
		}
		if _, err := bot.Send(message); err != nil {
			log.Println(err)
		}
		return
	}

	timeSet := fmt.Sprintln(userTime.Hour(), ":", userTime.Minute())
	userSetting, err := db.GetSettingById(update.Message.From.ID)
	if err != nil {
		panic(err)
	}
	userSetting.Time = timeSet
	_, err = db.UpdateSetting(*userSetting)
	if err != nil {
		panic(err)
	}

	//Обнавление крон задачи если влючены уведомления
	if userSetting.Notification {
		schedule.CreateNotificationsTasks(userSetting.ID, userSetting.Time)
	}

	msg := "Время задано"
	message := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
	message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Получать уведомления", CallbackData: &callBackWeatherNotifPush}},
			{{Text: "Установить время", CallbackData: &callBackWeatherNotifTime}},
			{{Text: "Отмена", CallbackData: &callBackWeatherMenu}},
		},
	}
	if _, err := bot.Send(message); err != nil {
		log.Println(err)
	}

	awaitTime[update.Message.From.ID] = false
}
