package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"

	basijibot "github.com/mavahedinia/BasijiCheckBot/BasijiBot"
	configs "github.com/mavahedinia/BasijiCheckBot/config"
	inmemory "github.com/mavahedinia/BasijiCheckBot/storage/in_memory"
)

func main() {
	config := configs.NewConfig()
	storage := inmemory.NewGroupInfoStorage(config)

	bot, err := tgbotapi.NewBotAPI(config.GetString("telegram.bot.token"))
	if err != nil {
		logrus.WithError(err).Panic("Failed to initialize bot")
	}
	basijiBot := basijibot.NewBasijiBot(config, bot, storage)

	logrus.WithField("botUserName", bot.Self.UserName).Info("Authorized on account")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			logrus.WithField("update", update).Info("unkown update received")
			continue
		}

		if handled := basijiBot.HandleNormalMessage(update.Message); handled && !update.Message.IsCommand() {
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		if msg, err := basijiBot.ValidateCommand(update.Message); err != nil {
			response := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
			response.ReplyToMessageID = update.Message.MessageID

			resp, err := bot.Send(response)
			if err != nil {
				logrus.WithError(err).WithField("response", resp).Info("error while replying.")
			}

			continue
		}

		logrus.WithFields(logrus.Fields{
			"command": update.Message.Command(),
			"user":    update.Message.From,
		}).Debug("Received a command")

		command := update.Message.Command()
		var msg string
		var err error
		switch command {
		case "basiji":
			msg, err = basijiBot.AddToListCommand(update.Message)
			break
		case "notbasiji":
			msg, err = basijiBot.RemoveFromListCommand(update.Message)
			break
		}

		if err != nil {
			logrus.WithError(err).Error("error occured in processing commands")
		}

		response := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
		response.ReplyToMessageID = update.Message.MessageID
		resp, err := bot.Send(response)
		if err != nil {
			logrus.WithError(err).WithField("response", resp).Info("error while replying.")
		}

	}
}
