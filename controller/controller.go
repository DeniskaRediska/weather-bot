package controller

import "github.com/go-telegram-bot-api/telegram-bot-api"

type handler struct {
	trigger func(update tgbotapi.Update) bool
	action  func(bot *tgbotapi.BotAPI, update tgbotapi.Update)
}

type BotController struct {
	Bot      *tgbotapi.BotAPI
	handlers []handler
}

func (c *BotController) When(trigger func(update tgbotapi.Update) bool, action func(bot *tgbotapi.BotAPI, update tgbotapi.Update)) {
	c.handlers = append(c.handlers, handler{trigger: trigger, action: action})
}

func (c BotController) HandleUpdate(update tgbotapi.Update) {
	for _, handler := range c.handlers {
		if handler.trigger(update) {
			handler.action(c.Bot, update)
			break
		}
	}
}
