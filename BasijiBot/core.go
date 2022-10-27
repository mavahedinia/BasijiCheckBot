package basijibot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/mavahedinia/BasijiCheckBot/storage"
)

type BasijiBot struct {
	config        *viper.Viper
	TgBot         *tgbotapi.BotAPI
	StorageDriver storage.GroupInfoInterface
}

func NewBasijiBot(config *viper.Viper, bot *tgbotapi.BotAPI, storageDriver storage.GroupInfoInterface) *BasijiBot {
	return &BasijiBot{
		config:        config,
		TgBot:         bot,
		StorageDriver: storageDriver,
	}
}

func (bb *BasijiBot) ValidateCommand(message *tgbotapi.Message) (string, error) {
	if message.ReplyToMessage == nil {
		return "لطفا به فرد بسیجی ریپلای بزنید سپس بات را صدا کنید.", &BasijiBotError{msg: "NoReply"}
	}

	command := message.Command()
	if command == "basiji" {
		return "", nil
	}

	if command == "notbasiji" {
		return "", nil
	}

	return "دستور اشتباه استفاده شده است. برای اضافه کردن بسیجی به بات، از فرمان /basiji و برای تصحیح خطا از فرمان /notbasiji حین ریپلای به فرد مورد نظر استفاده کنید.", &BasijiBotError{msg: "InvalidCommand"}
}

func (bb *BasijiBot) AddToListCommand(message *tgbotapi.Message) (string, error) {
	if message.ReplyToMessage.From.ID == bb.TgBot.Self.ID {
		return "آقا ما خودی هستیم :(", &BasijiBotError{msg: "NotMeBruh"}
	}

	if msg, err := bb.hasPriviledge(message.Chat, message.From); err != nil {
		return msg, err
	}

	err := bb.StorageDriver.AddBasiji(message.Chat, message.ReplyToMessage.From)
	if err != nil {
		return "شناسایی بسیجی با مشکل مواجه شد :(", err
	}

	return "بسیجی با موفقیت به لیست بسیجی‌ها اضافه شد!", nil
}

func (bb *BasijiBot) RemoveFromListCommand(message *tgbotapi.Message) (string, error) {
	if msg, err := bb.hasPriviledge(message.Chat, message.From); err != nil {
		return msg, err
	}

	err := bb.StorageDriver.RemoveBasiji(message.Chat, message.ReplyToMessage.From)
	if err != nil {
		return "مشکلی در بات پیش آمده و عملیات موفقیت آمیز نبود :(", err
	}

	return "فرد مورد نظر از لیست بسیجی‌ها خارج شد! :)", nil
}

func (bb *BasijiBot) HandleNormalMessage(message *tgbotapi.Message) bool {
	if isBasiji, _ := bb.StorageDriver.IsBasiji(message.Chat, message.From); !isBasiji {
		return false
	}

	response := tgbotapi.NewPhotoUpload(message.Chat.ID, "photos/sandis.jpg")
	response.ReplyToMessageID = message.MessageID
	response.Caption = "ساندیس بقول :*"

	resp, err := bb.TgBot.Send(response)
	if err != nil {
		logrus.WithError(err).WithField("response", resp).Info("error while replying.")
	}

	return true
}

func (bb *BasijiBot) hasPriviledge(chat *tgbotapi.Chat, user *tgbotapi.User) (string, error) {
	chatMember, err := bb.TgBot.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID:             chat.ID,
		UserID:             user.ID,
		SuperGroupUsername: chat.ChatConfig().SuperGroupUsername,
	})
	if err != nil {
		logrus.WithError(err).Error("failed to check if user is chat admin or not.")
	}

	if !chatMember.IsAdministrator() && !chatMember.IsCreator() {
		return "فقط ادمین‌های چت امکان استفاده از این بات را دارند!", &BasijiBotError{msg: "NotAdmin"}
	}

	return "", nil
}

type BasijiBotError struct {
	msg string
}

func (bbe *BasijiBotError) Error() string {
	return bbe.msg
}
