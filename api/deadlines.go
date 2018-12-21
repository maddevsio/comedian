package api

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (ba *BotAPI) addTime(accessLevel int, channelID, params string) string {
	if accessLevel > 3 {
		return ba.Bot.Translate.AccessAtLeastPM
	}

	timeInt, err := utils.ParseTimeTextToInt(params)
	if err != nil {
		return err.Error()
	}
	err = ba.Bot.DB.CreateStandupTime(timeInt, channelID)
	if err != nil {
		logrus.Errorf("BotAPI: CreateStandupTime failed: %v\n", err)
		return ba.Bot.Translate.SomethingWentWrong
	}
	channelMembers, err := ba.Bot.DB.ListChannelMembers(channelID)
	if err != nil {
		logrus.Errorf("BotAPI: ListChannelMembers failed: %v\n", err)
	}
	if len(channelMembers) == 0 {
		return fmt.Sprintf(ba.Bot.Translate.AddStandupTimeNoUsers, timeInt)
	}
	return fmt.Sprintf(ba.Bot.Translate.AddStandupTime, timeInt)
}

func (ba *BotAPI) removeTime(accessLevel int, channelID string) string {
	if accessLevel > 3 {
		return ba.Bot.Translate.AccessAtLeastPM
	}
	err := ba.Bot.DB.DeleteStandupTime(channelID)
	if err != nil {
		logrus.Errorf("BotAPI: DeleteStandupTime failed: %v\n", err)
		return ba.Bot.Translate.SomethingWentWrong
	}
	st, err := ba.Bot.DB.ListChannelMembers(channelID)
	if len(st) != 0 {
		return ba.Bot.Translate.RemoveStandupTimeWithUsers
	}
	return fmt.Sprintf(ba.Bot.Translate.RemoveStandupTime)
}

func (ba *BotAPI) showTime(channelID string) string {
	standupTime, err := ba.Bot.DB.GetChannelStandupTime(channelID)
	if err != nil || standupTime == int64(0) {
		logrus.Errorf("GetChannelStandupTime failed: %v", err)
		return ba.Bot.Translate.ShowNoStandupTime
	}
	return fmt.Sprintf(ba.Bot.Translate.ShowStandupTime, standupTime)
}
