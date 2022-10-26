package inmemory

import (
	"encoding/json"
	"io/ioutil"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type MemberInfo struct {
	User     *tgbotapi.User
	IsBasiji bool
}

func NewMemberInfo(user *tgbotapi.User, IsBasiji bool) *MemberInfo {
	return &MemberInfo{
		User:     user,
		IsBasiji: IsBasiji,
	}
}

type ChatInfo struct {
	Chat          *tgbotapi.Chat
	BasijiMembers map[int]*MemberInfo
}

func NewChatInfo(chat *tgbotapi.Chat, bms map[int]*MemberInfo) *ChatInfo {
	if bms == nil {
		bms = make(map[int]*MemberInfo)
	}

	return &ChatInfo{
		Chat:          chat,
		BasijiMembers: bms,
	}
}

type GroupInfoStorage struct {
	config *viper.Viper
	groups map[int64]*ChatInfo
}

func NewGroupInfoStorage(config *viper.Viper) *GroupInfoStorage {
	gis := &GroupInfoStorage{
		config: config,
		groups: make(map[int64]*ChatInfo),
	}

	gis.loadData()

	go gis.persistData()

	return gis
}

func (gis *GroupInfoStorage) persistData() {
	sleepDuration := gis.config.GetDuration("storage.wait-time")
	for true {
		time.Sleep(sleepDuration)

		gis.writeData()
	}
}

func (gis *GroupInfoStorage) loadData() {
	filename := gis.config.GetString("storage.filename")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		logrus.WithError(err).Error("error in opening data file.")
		return
		// panic(1)
	}

	err = json.Unmarshal(data, &gis.groups)
	if err != nil {
		logrus.WithError(err).Error("error in reading data.")
	}
}

func (gis *GroupInfoStorage) writeData() {
	filename := gis.config.GetString("storage.filename")
	logrus.WithField("filename", filename).Info("writing to file.")

	content, err := json.Marshal(gis.groups)
	if err != nil {
		logrus.WithError(err).Error("error in marshalling groups data.")
		return
	}

	err = ioutil.WriteFile(filename, content, 0o644)
	if err != nil {
		logrus.WithError(err).Error("error in writing data.")
	}
}

func (gis *GroupInfoStorage) AddBasiji(chat *tgbotapi.Chat, user *tgbotapi.User) error {
	chatInfo, ok := gis.groups[chat.ID]
	if !ok {
		gis.groups[chat.ID] = NewChatInfo(chat, nil)
		chatInfo, _ = gis.groups[chat.ID]
	}

	_, exists := chatInfo.BasijiMembers[user.ID]
	if !exists {
		chatInfo.BasijiMembers[user.ID] = NewMemberInfo(user, true)
		return nil
	}

	mi, _ := chatInfo.BasijiMembers[user.ID]
	mi.IsBasiji = true

	return nil
}

func (gis *GroupInfoStorage) RemoveBasiji(chat *tgbotapi.Chat, user *tgbotapi.User) error {
	chatInfo, ok := gis.groups[chat.ID]
	if !ok { // Group not exists in db
		return nil
	}

	_, exists := chatInfo.BasijiMembers[user.ID]
	if !exists { // User not exists is basijis list
		return nil
	}

	mi, _ := chatInfo.BasijiMembers[user.ID]
	mi.IsBasiji = false

	return nil
}

func (gis *GroupInfoStorage) IsBasiji(chat *tgbotapi.Chat, user *tgbotapi.User) (bool, error) {
	chatInfo, ok := gis.groups[chat.ID]
	if !ok { // Group not exists in db
		return false, nil
	}

	_, exists := chatInfo.BasijiMembers[user.ID]
	if !exists { // User not exists is basijis list
		return false, nil
	}

	mi, _ := chatInfo.BasijiMembers[user.ID]
	if mi.IsBasiji {
		return true, nil
	}

	return false, nil
}
