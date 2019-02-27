package botuser

import (
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
	"github.com/olebedev/when/rules/ru"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/translation"
)

func (bot *Bot) addTime(accessLevel int, channelID, params string) string {
	if accessLevel > pmAccess {
		accessAtLeastPM, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AccessAtLeastPM", 0, nil)
		return accessAtLeastPM
	}

	w := when.New(nil)
	w.Add(en.All...)
	w.Add(ru.All...)

	somethingWentWrong, _ := translation.Translate(bot.bundle, bot.Properties.Language, "SomethingWentWrong", 0, nil)

	r, err := w.Parse(params, time.Now())
	if err != nil {
		return somethingWentWrong
	}
	if r == nil {
		return somethingWentWrong
	}

	err = bot.db.CreateStandupTime(r.Time.Unix(), channelID)
	if err != nil {
		return somethingWentWrong
	}

	channelMembers, err := bot.db.ListChannelMembers(channelID)
	if err != nil {
		log.Errorf("BotAPI: ListChannelMembers failed: %v\n", err)
	}

	if len(channelMembers) == 0 {
		addStandupTimeNoUsers, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AddStandupTimeNoUsers", 0, map[string]interface{}{"timeInt": r.Time.Unix()})
		return addStandupTimeNoUsers
	}

	addStandupTime, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AddStandupTime", 0, map[string]interface{}{"timeInt": r.Time.Unix()})
	return addStandupTime
}

func (bot *Bot) removeTime(accessLevel int, channelID string) string {
	if accessLevel > pmAccess {
		accessAtLeastPM, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AccessAtLeastPM", 0, nil)
		return accessAtLeastPM
	}

	err := bot.db.DeleteStandupTime(channelID)
	if err != nil {
		somethingWentWrong, _ := translation.Translate(bot.bundle, bot.Properties.Language, "SomethingWentWrong", 0, nil)
		return somethingWentWrong

	}
	st, err := bot.db.ListChannelMembers(channelID)
	if len(st) != 0 {
		removeStandupTimeWithUsers, _ := translation.Translate(bot.bundle, bot.Properties.Language, "RemoveStandupTimeWithUsers", 0, nil)
		return removeStandupTimeWithUsers
	}

	removeStandupTime, _ := translation.Translate(bot.bundle, bot.Properties.Language, "RemoveStandupTime", 0, nil)
	return removeStandupTime
}

func (bot *Bot) showTime(channelID string) string {
	standupTime, err := bot.db.GetChannelStandupTime(channelID)
	if err != nil || standupTime == int64(0) {
		showNoStandupTime, _ := translation.Translate(bot.bundle, bot.Properties.Language, "ShowNoStandupTime", 0, nil)
		return showNoStandupTime
	}

	showStandupTime, _ := translation.Translate(bot.bundle, bot.Properties.Language, "ShowStandupTime", 0, map[string]interface{}{"standuptime": standupTime})
	return showStandupTime
}
