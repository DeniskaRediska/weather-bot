package controller

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

type handleFunc func(bot *tgbotapi.BotAPI, update tgbotapi.Update) error

type handler struct {
	trigger func(update tgbotapi.Update) bool
	action  handleFunc
}

type BotController struct {
	Bot      *tgbotapi.BotAPI
	handlers []handler
}

func (c *BotController) When(trigger func(update tgbotapi.Update) bool, action handleFunc) {
	c.handlers = append(c.handlers, handler{trigger: trigger, action: action})
}

func (c BotController) HandleUpdate(update tgbotapi.Update) {
	for _, handler := range c.handlers {
		if handler.trigger(update) {
			err := handler.action(c.Bot, update)
			if err != nil {
				log.Println(err)

				var idUser int
				if update.Message != nil {
					idUser = update.Message.From.ID
				} else if update.CallbackQuery != nil {
					idUser = update.CallbackQuery.From.ID
				}
				if idUser != 0 {
					msg := tgbotapi.NewMessage(int64(idUser), "ops, you have some trouble")
					_, err := c.Bot.Send(msg)
					if err != nil {
						log.Println(err)
					}
				}
			}
			break
		}
	}
}
