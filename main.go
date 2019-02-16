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

	cancelKeyboard = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Отмена", CallbackData: &callBackWeatherMenu}},
		},
	}
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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		botController.HandleUpdate(update)
	}
}

func findOrCreateSetting(id int) (*db.SettingBot, error) {
	settings, err := db.GetSettingById(id)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		settings = &db.SettingBot{
			ID:           id,
			Lon:          0.0,
			Lat:          0.0,
			Notification: false,
			Time:         "9:00",
		}
		_, err := db.InsertSetting(*settings)
		if err != nil {
			return nil, err
		}
	}
	return settings, nil
}

func createMainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Погода в данный момент", CallbackData: &callBackWeatherRightNow}},
			{{Text: "Настройка уведомлений", CallbackData: &callBackWeatherNotification}},
			{{Text: "Задать локацию", CallbackData: &callBackWeatherLocation}},
		},
	}
}

func createSecondSettingKeyboard(setting *db.SettingBot) tgbotapi.InlineKeyboardMarkup {
	msg := "уведомляшки"
	if setting.Notification {
		msg = "отключить уведомления"
	} else {
		msg = "включить уведомления"
	}
	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: msg, CallbackData: &callBackWeatherNotifPush}},
			{{Text: "Установить время", CallbackData: &callBackWeatherNotifTime}},
			{{Text: "Отмена", CallbackData: &callBackWeatherMenu}},
		},
	}
}

func getSettingInfo(setting *db.SettingBot) string {
	msg := "Настройка уведомлений"
	if setting.Notification {
		msg = fmt.Sprintln(msg, "\nУведомления включены")
	} else {
		msg = fmt.Sprintln(msg, "\nУведомления отключены")
	}
	msg = fmt.Sprintf("%sТекущее время: %s", msg, setting.Time)
	return msg
}

func updateMessageBot(bot *tgbotapi.BotAPI, chatId int64, msgId int, textMsg string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	if keyboard != nil && textMsg == "" {
		message := tgbotapi.NewEditMessageReplyMarkup(chatId, msgId, *keyboard)
		if _, err := bot.Send(message); err != nil {
			return err
		}
	}
	if textMsg != "" {
		editText := tgbotapi.NewEditMessageText(chatId, msgId, textMsg)
		editText.ReplyMarkup = keyboard
		if _, err := bot.Send(editText); err != nil {
			return err
		}
	}
	return nil
}

func isStart(update tgbotapi.Update) bool {
	isStart := update.Message != nil && update.Message.Text == "/start"
	return isStart
}
func handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {

	_, err := findOrCreateSetting(update.Message.From.ID)
	if err != nil {
		return err
	}

	message := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет, я погодный бот")
	message.ReplyMarkup = createMainKeyboard()
	if _, err := bot.Send(message); err != nil {
		return err
	}
	return nil
}

//
func isLocationButton(update tgbotapi.Update) bool {
	isLocationButton := update.CallbackQuery != nil && update.CallbackQuery.Data == callBackWeatherLocation
	return isLocationButton
}
func handleLocation(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пожалуйста, отправьте мне свою локацию\n\"Прикрепить\"->\"Location\"")
	message.ReplyMarkup = cancelKeyboard
	if _, err := bot.Send(message); err != nil {
		return err
	}
	awaitLocation[update.CallbackQuery.From.ID] = true
	return nil
}

//
func isAwaitLocationYes(update tgbotapi.Update) bool {
	isAwaitLocation := update.Message != nil && awaitLocation[update.Message.From.ID] == true && update.Message.Location != nil
	return isAwaitLocation
}
func handleAwaitYes(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	setting, err := findOrCreateSetting(update.Message.From.ID)
	if err != nil {
		return err
	}
	setting.Lon = update.Message.Location.Longitude
	setting.Lat = update.Message.Location.Latitude
	_, err = db.UpdateSetting(*setting)
	if err != nil {
		return err
	}
	awaitLocation[update.Message.From.ID] = false

	message := tgbotapi.NewMessage(update.Message.Chat.ID, "Локация успешно задана")
	message.ReplyMarkup = createMainKeyboard()
	if _, err := bot.Send(message); err != nil {
		return err
	}
	return nil
}

//
func isAwaitLocationNo(update tgbotapi.Update) bool {
	isAwaitLocation := update.Message != nil && awaitLocation[update.Message.From.ID] == true && update.Message.Location == nil
	return isAwaitLocation
}
func handleAwaitNo(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	message := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, отправьте мне локацию\n\"Прикрепить\"->\"Location\"")
	message.ReplyMarkup = cancelKeyboard
	if _, err := bot.Send(message); err != nil {
		return err
	}
	return nil
}

//
func isToMenuButton(update tgbotapi.Update) bool {
	backToMenu := update.CallbackQuery != nil && update.CallbackQuery.Data == callBackWeatherMenu
	return backToMenu
}
func backToMenu(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	awaitTime[update.CallbackQuery.From.ID] = false
	awaitLocation[update.CallbackQuery.From.ID] = false

	message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Отмена. Возврат в меню")
	message.ReplyMarkup = createMainKeyboard()
	if _, err := bot.Send(message); err != nil {
		return err
	}
	return nil
}

//
func isTempRightNowButton(update tgbotapi.Update) bool {
	TempRightNow := update.CallbackQuery != nil && update.CallbackQuery.Data == callBackWeatherRightNow
	return TempRightNow
}
func handleTempRightNow(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	setting, err := findOrCreateSetting(update.CallbackQuery.From.ID)
	if err != nil {
		return err
	}

	if setting.Lat == 0 && setting.Lon == 0 {
		message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пожалуйста, отправьте мне свою локацию\n\"Прикрепить\"->\"Location\"")
		message.ReplyMarkup = cancelKeyboard
		if _, err := bot.Send(message); err != nil {
			return err
		}
		awaitLocation[update.CallbackQuery.From.ID] = true
	} else {
		cityName, temp, err := openweather.GetCurrentWeather(setting.Lat, setting.Lon)
		if err != nil {
			return err
		}
		msg := fmt.Sprintln("Ваш город - ", cityName, "\nТемпература - ", temp)
		message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, msg)
		message.ReplyMarkup = createMainKeyboard()
		if _, err := bot.Send(message); err != nil {
			return err
		}
	}
	return nil
}

//
func isNotificationButton(update tgbotapi.Update) bool {
	isNotificationButton := update.CallbackQuery != nil && update.CallbackQuery.Data == callBackWeatherNotification
	return isNotificationButton
}
func handleNotificationButton(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	setting, err := findOrCreateSetting(update.CallbackQuery.From.ID)
	if err != nil {
		return err
	}
	msg := getSettingInfo(setting)
	message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, msg)
	message.ReplyMarkup = createSecondSettingKeyboard(setting)
	if _, err := bot.Send(message); err != nil {
		log.Println(err)
	}
	return nil
}

//
func isNotifSetting(update tgbotapi.Update) bool {
	isNotif := update.CallbackQuery != nil && update.CallbackQuery.Data == callBackWeatherNotifPush
	return isNotif
}
func handleNotifPush(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	setting, err := findOrCreateSetting(update.CallbackQuery.From.ID)
	if err != nil {
		return err
	}
	var msg string
	if setting.Notification {
		setting.Notification = false
		_, err = db.UpdateSetting(*setting)
		if err != nil {
			return err
		}
		schedule.DeleteNotificationsTask(setting.ID)
		msg = fmt.Sprintln("Уведомления были отключены")
	} else {
		setting.Notification = true
		_, err = db.UpdateSetting(*setting)
		if err != nil {
			return err
		}
		schedule.CreateNotificationsTasks(setting.ID, setting.Time)
		msg = fmt.Sprintln("Уведомления были включены")
	}
	markup := createSecondSettingKeyboard(setting)
	return updateMessageBot(bot,
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		msg,
		&markup,
	)
}

//
func isTimeSetting(update tgbotapi.Update) bool {
	isTimeSetting := update.CallbackQuery != nil && update.CallbackQuery.Data == callBackWeatherNotifTime
	return isTimeSetting
}
func handleTimeSetting(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	message := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пожалуйста, отправьте мне время в формате XX:XX")
	message.ReplyMarkup = cancelKeyboard
	if _, err := bot.Send(message); err != nil {
		return err
	}
	awaitTime[update.CallbackQuery.From.ID] = true
	return nil
}

//
func isAwaitTime(update tgbotapi.Update) bool {
	isAwaitTime := update.Message != nil && awaitTime[update.Message.From.ID] == true
	return isAwaitTime
}
func handleAwaitTime(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	userTime, err := time.Parse("15:04", update.Message.Text)
	if err != nil {
		message := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, отправьтьте время в формате\nXX:XX")
		message.ReplyMarkup = cancelKeyboard
		if _, err := bot.Send(message); err != nil {
			return err
		}
		return nil
	}

	timeSet := fmt.Sprintf("%d:%d", userTime.Minute(), userTime.Hour())
	userSetting, err := findOrCreateSetting(update.Message.From.ID)
	if err != nil {
		return err
	}
	userSetting.Time = timeSet
	_, err = db.UpdateSetting(*userSetting)
	if err != nil {
		return err
	}

	//Обнавление крон задачи если влючены уведомления
	if userSetting.Notification {
		schedule.CreateNotificationsTasks(userSetting.ID, userSetting.Time)
	}
	awaitTime[update.Message.From.ID] = false
	msg := "Время задано"
	message := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
	message.ReplyMarkup = createSecondSettingKeyboard(userSetting)
	if _, err := bot.Send(message); err != nil {
		return err
	}
	return nil
}
