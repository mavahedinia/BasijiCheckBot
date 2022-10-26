package storage

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type GroupInfoInterface interface {
	AddBasiji(Chat *tgbotapi.Chat, User *tgbotapi.User) error
	RemoveBasiji(Chat *tgbotapi.Chat, User *tgbotapi.User) error
	IsBasiji(Chat *tgbotapi.Chat, User *tgbotapi.User) (bool, error)
}
