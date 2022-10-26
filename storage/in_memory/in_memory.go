package inmemory

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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
	groups map[int64]*ChatInfo
}

func NewGroupInfoStorage() *GroupInfoStorage {
	return &GroupInfoStorage{
		groups: make(map[int64]*ChatInfo),
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
