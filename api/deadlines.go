package api

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (r *REST) addTime(accessLevel int, channelID, params string) string {
	if accessLevel > 3 {
		return r.slack.Translate.AccessAtLeastPM
	}

	timeInt, err := utils.ParseTimeTextToInt(params)
	if err != nil {
		return err.Error()
	}
	err = r.db.CreateStandupTime(timeInt, channelID)
	if err != nil {
		logrus.Errorf("rest: CreateStandupTime failed: %v\n", err)
		return r.slack.Translate.SomethingWentWrong
	}
	channelMembers, err := r.db.ListChannelMembers(channelID)
	if err != nil {
		logrus.Errorf("rest: ListChannelMembers failed: %v\n", err)
	}
	if len(channelMembers) == 0 {
		return fmt.Sprintf(r.slack.Translate.AddStandupTimeNoUsers, timeInt)
	}
	return fmt.Sprintf(r.slack.Translate.AddStandupTime, timeInt)
}

func (r *REST) removeTime(accessLevel int, channelID string) string {
	if accessLevel > 3 {
		return r.slack.Translate.AccessAtLeastPM
	}
	err := r.db.DeleteStandupTime(channelID)
	if err != nil {
		logrus.Errorf("rest: DeleteStandupTime failed: %v\n", err)
		return r.slack.Translate.SomethingWentWrong
	}
	st, err := r.db.ListChannelMembers(channelID)
	if len(st) != 0 {
		return r.slack.Translate.RemoveStandupTimeWithUsers
	}
	return fmt.Sprintf(r.slack.Translate.RemoveStandupTime)
}

func (r *REST) showTime(channelID string) string {
	standupTime, err := r.db.GetChannelStandupTime(channelID)
	if err != nil || standupTime == int64(0) {
		logrus.Errorf("GetChannelStandupTime failed: %v", err)
		return r.slack.Translate.ShowNoStandupTime
	}
	return fmt.Sprintf(r.slack.Translate.ShowStandupTime, standupTime)
}
