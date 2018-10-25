package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/reporting"
	"github.com/maddevsio/comedian/storage"
	"github.com/maddevsio/comedian/teammonitoring"
	"github.com/maddevsio/comedian/utils"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

// REST struct used to handle slack requests (slash commands)
type REST struct {
	db      storage.Storage
	echo    *echo.Echo
	conf    config.Config
	decoder *schema.Decoder
	report  *reporting.Reporter
	slack   *chat.Slack
	api     *slack.Client
}

const (
	commandAdd    = "/add"
	commandDelete = "/delete"
	commandList   = "/list"

	commandAddTime    = "/standup_time_set"
	commandRemoveTime = "/standup_time_remove"
	commandListTime   = "/standup_time"

	commandAddTimeTable    = "/timetable_set"
	commandRemoveTimeTable = "/timetable_remove"
	commandShowTimeTable   = "/timetable_show"

	commandReportByProject       = "/report_by_project"
	commandReportByUser          = "/report_by_user"
	commandReportByUserInProject = "/report_by_user_in_project"

	commandHelp = "/helper"
)

//ResponseText is Comedian API response text message to be displayed
var ResponseText string

// NewRESTAPI creates API for Slack commands
func NewRESTAPI(slack *chat.Slack) (*REST, error) {
	e := echo.New()
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	rep := reporting.NewReporter(slack)

	r := &REST{
		echo:    e,
		decoder: decoder,
		report:  rep,
		db:      slack.DB,
		slack:   slack,
		api:     slack.API,
		conf:    slack.Conf,
	}

	r.initEndpoints()
	return r, nil
}

func (r *REST) initEndpoints() {
	endPoint := fmt.Sprintf("/commands%s", r.conf.SecretToken)
	r.echo.POST(endPoint, r.handleCommands)
}

// Start starts http server
func (r *REST) Start() error {
	return r.echo.Start(r.conf.HTTPBindAddr)
}

func (r *REST) handleCommands(c echo.Context) error {
	form, err := c.FormParams()
	if err != nil {
		logrus.Errorf("rest: c.FormParams failed: %v\n", err)
	}
	command := form.Get("command")

	switch command {
	case commandHelp:
		return r.helpCommand(c, form)
	case commandAdd:
		return r.addCommand(c, form)
	case commandList:
		return r.listCommand(c, form)
	case commandDelete:
		return r.deleteCommand(c, form)
	case commandAddTime:
		return r.addTime(c, form)
	case commandRemoveTime:
		return r.removeTime(c, form)
	case commandListTime:
		return r.listTime(c, form)
	case commandAddTimeTable:
		return r.addTimeTable(c, form)
	case commandRemoveTimeTable:
		return r.removeTimeTable(c, form)
	case commandShowTimeTable:
		return r.showTimeTable(c, form)
	case commandReportByProject:
		return r.reportByProject(c, form)
	case commandReportByUser:
		return r.reportByUser(c, form)
	case commandReportByUserInProject:
		return r.reportByProjectAndUser(c, form)
	default:
		return c.String(http.StatusNotImplemented, "Not implemented")
	}
}

func (r *REST) helpCommand(c echo.Context, f url.Values) error {
	_, _, _, _, err := r.processCommand(c, f)
	if err != nil {
		return c.String(http.StatusOK, r.conf.Translate.SomethingWentWrong)
	}
	return c.String(http.StatusOK, r.conf.Translate.HelpCommand)
}

func (r *REST) addCommand(c echo.Context, f url.Values) error {
	users, role, channel, accessLevel, err := r.processCommand(c, f)
	if err != nil {
		return c.String(http.StatusOK, r.conf.Translate.SomethingWentWrong)
	}
	switch role {
	case "admin", "админ":
		if accessLevel > 2 {
			return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastAdmin)
		}
		return c.String(http.StatusOK, r.addAdmins(users))
	case "developer", "разработчик", "":
		if accessLevel > 3 {
			return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastPM)
		}
		return c.String(http.StatusOK, r.addUsers(users, channel))
	case "pm", "пм":
		if accessLevel > 2 {
			return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastAdmin)
		}
		return c.String(http.StatusOK, r.addPMs(users, channel))
	default:
		return c.String(http.StatusOK, r.conf.Translate.NeedCorrectUserRole)
	}
}

func (r *REST) listCommand(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}

	switch ca.Text {
	case "admin", "админ":
		return c.String(http.StatusOK, r.listAdmins())
	case "developer", "разработчик", "":
		return c.String(http.StatusOK, r.listUsers(ca.ChannelID))
	case "pm", "пм":
		return c.String(http.StatusOK, r.listPMs(ca.ChannelID))
	default:
		return c.String(http.StatusOK, r.conf.Translate.NeedCorrectUserRole)
	}

}

func (r *REST) deleteCommand(c echo.Context, f url.Values) error {

	users, role, channel, accessLevel, err := r.processCommand(c, f)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	switch role {
	case "admin", "админ":
		if accessLevel > 2 {
			return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastAdmin)
		}
		return c.String(http.StatusOK, r.deleteAdmins(users))
	case "developer", "разработчик":
		if accessLevel > 3 {
			return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastPM)
		}
		return c.String(http.StatusOK, r.deleteUsers(users, channel))
	case "pm", "пм":
		if accessLevel > 2 {
			return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastAdmin)
		}
		return c.String(http.StatusOK, r.deletePMs(users, channel))
	default:
		return c.String(http.StatusOK, r.conf.Translate.NeedCorrectUserRole)
	}
}

func (r *REST) addUsers(users []string, channel string) string {
	var failed, exist, added, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := r.db.FindChannelMemberByUserID(userID, channel)
		if err != nil {
			logrus.Errorf("Rest FindChannelMemberByUserID failed: %v", err)
			chanMember, _ := r.db.CreateChannelMember(model.ChannelMember{
				UserID:    userID,
				ChannelID: channel,
			})
			logrus.Infof("ChannelMember created! ID:%v", chanMember.ID)
		}
		if user.UserID == userID && user.ChannelID == channel {
			exist += u
			continue
		}
		added += u
	}

	if len(failed) != 0 {
		text += fmt.Sprintf(r.conf.Translate.AddUsersFailed, failed)
	}
	if len(exist) != 0 {
		text += fmt.Sprintf(r.conf.Translate.AddUsersExist, exist)
	}
	if len(added) != 0 {
		text += fmt.Sprintf(r.conf.Translate.AddUsersAdded, added)
	}
	return text
}

func (r *REST) addPMs(users []string, channel string) string {
	var failed, exist, added, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		_, err := r.db.FindChannelMemberByUserID(userID, channel)
		if err != nil {
			r.db.CreatePM(model.ChannelMember{
				UserID:    userID,
				ChannelID: channel,
			})
			added += u
			continue
		}
		exist += u
	}
	if len(failed) != 0 {
		text += fmt.Sprintf(r.conf.Translate.AddPMsFailed, failed)
	}
	if len(exist) != 0 {
		text += fmt.Sprintf(r.conf.Translate.AddPMsExist, exist)
	}
	if len(added) != 0 {
		text += fmt.Sprintf(r.conf.Translate.AddPMsAdded, added)
	}

	return text
}

func (r *REST) addAdmins(users []string) string {
	var failed, exist, added, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := r.db.SelectUser(userID)
		if err != nil {
			failed += u
			continue
		}
		if user.Role == "admin" {
			exist += u
			continue
		}
		user.Role = "admin"
		r.db.UpdateUser(user)
		message := r.conf.Translate.PMAssigned
		err = r.slack.SendUserMessage(userID, message)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		added += u
	}

	if len(failed) != 0 {
		text += fmt.Sprintf(r.conf.Translate.AddAdminsFailed, failed)
	}
	if len(exist) != 0 {
		text += fmt.Sprintf(r.conf.Translate.AddAdminsExist, exist)
	}
	if len(added) != 0 {
		text += fmt.Sprintf(r.conf.Translate.AddAdminsAdded, added)
	}

	return text
}

func (r *REST) listUsers(channel string) string {
	users, err := r.db.ListChannelMembers(channel)
	if err != nil {
		return fmt.Sprintf("failed to list users :%v\n", err)
	}
	var userIDs []string
	for _, user := range users {
		userIDs = append(userIDs, "<@"+user.UserID+">")
	}
	if len(userIDs) < 1 {
		return r.conf.Translate.ListNoStandupers
	}
	return fmt.Sprintf(r.conf.Translate.ListStandupers, strings.Join(userIDs, ", "))
}

func (r *REST) listAdmins() string {
	admins, err := r.db.ListAdmins()
	if err != nil {
		return fmt.Sprintf("failed to list users :%v\n", err)
	}
	var userNames []string
	for _, admin := range admins {
		userNames = append(userNames, "<@"+admin.UserName+">")
	}
	if len(userNames) < 1 {
		return r.conf.Translate.ListNoAdmins
	}
	return fmt.Sprintf(r.conf.Translate.ListAdmins, strings.Join(userNames, ", "))

}

func (r *REST) listPMs(channel string) string {
	users, err := r.db.ListPMs(channel)
	if err != nil {
		return fmt.Sprintf("failed to list users :%v\n", err)
	}
	var userIDs []string
	for _, user := range users {
		userIDs = append(userIDs, "<@"+user.UserID+">")
	}
	if len(userIDs) < 1 {
		return r.conf.Translate.ListNoPMs
	}
	return fmt.Sprintf(r.conf.Translate.ListPMs, strings.Join(userIDs, ", "))
}

func (r *REST) deleteUsers(users []string, channel string) string {
	var failed, deleted, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := r.db.FindChannelMemberByUserID(userID, channel)
		if err != nil {
			logrus.Errorf("rest: FindChannelMemberByUserID failed: %v\n", err)
			failed += u
			continue
		}
		r.db.DeleteChannelMember(user.UserID, channel)
		deleted += u
	}

	if len(failed) != 0 {
		text += fmt.Sprintf("Could not remove the following users as developers: %v\n", failed)
	}
	if len(deleted) != 0 {
		text += fmt.Sprintf("The following users were removed as developers: %v\n", deleted)
	}

	return text
}

func (r *REST) deletePMs(users []string, channel string) string {
	var failed, deleted, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, userName := utils.SplitUser(u)
		user, err := r.db.FindChannelMemberByUserID(userID, channel)
		if err != nil {
			logrus.Errorf("rest: FindChannelMemberByUserID failed: %v\n", err)
			failed += u
			continue
		}
		if user.RoleInChannel != "PM" {
			logrus.Errorf("rest: User %v is not PM in %v! Skip\n", userName, user.ChannelID)
			failed += u
			continue
		}
		r.db.DeleteChannelMember(user.UserID, channel)
		deleted += u
	}
	if len(failed) != 0 {
		text += fmt.Sprintf("Could not remove users as PMs: %v\n", failed)
	}
	if len(deleted) != 0 {
		text += fmt.Sprintf("Users are removed as PMs: %v\n", deleted)
	}
	return text
}

func (r *REST) deleteAdmins(users []string) string {
	var failed, deleted, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := r.db.SelectUser(userID)
		if err != nil {
			failed += u
			continue
		}
		if user.Role != "admin" {
			failed += u
			continue
		}
		user.Role = ""
		r.db.UpdateUser(user)
		message := fmt.Sprintf(r.conf.Translate.PMRemoved)
		err = r.slack.SendUserMessage(userID, message)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		deleted += u
	}

	if len(failed) != 0 {
		text += fmt.Sprintf("Could not remove users as admins: %v\n", failed)
	}
	if len(deleted) != 0 {
		text += fmt.Sprintf("Users are removed as admins: %v\n", deleted)
	}

	return text
}

func (r *REST) addTime(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}

	accessLevel, _ := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	logrus.Infof("Access level for %v in %v is %v", f.Get("user_id"), f.Get("channel_id"), accessLevel)
	if accessLevel > 3 {
		return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastPM)
	}

	timeInt, err := utils.ParseTimeTextToInt(ca.Text)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	err = r.db.CreateStandupTime(timeInt, ca.ChannelID)
	if err != nil {
		logrus.Errorf("rest: CreateStandupTime failed: %v\n", err)
		return c.String(http.StatusOK, r.conf.Translate.SomethingWentWrong)
	}
	channelMembers, err := r.db.ListChannelMembers(ca.ChannelID)
	if err != nil {
		logrus.Errorf("rest: ListChannelMembers failed: %v\n", err)
	}
	if len(channelMembers) == 0 {
		return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.AddStandupTimeNoUsers, timeInt))
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.AddStandupTime, timeInt))
}

func (r *REST) removeTime(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}

	accessLevel, _ := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	logrus.Infof("Access level for %v in %v is %v", f.Get("user_id"), f.Get("channel_id"), accessLevel)
	if accessLevel > 3 {
		return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastPM)
	}

	err = r.db.DeleteStandupTime(ca.ChannelID)
	if err != nil {
		logrus.Errorf("rest: DeleteStandupTime failed: %v\n", err)
		return c.String(http.StatusOK, r.conf.Translate.SomethingWentWrong)
	}
	st, err := r.db.ListChannelMembers(ca.ChannelID)
	if len(st) != 0 {
		return c.String(http.StatusOK, r.conf.Translate.RemoveStandupTimeWithUsers)
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.RemoveStandupTime, ca.ChannelName))
}

func (r *REST) listTime(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}

	standupTime, err := r.db.GetChannelStandupTime(ca.ChannelID)
	if err != nil || standupTime == int64(0) {
		logrus.Errorf("GetChannelStandupTime failed: %v", err)
		return c.String(http.StatusOK, r.conf.Translate.ShowNoStandupTime)
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.ShowStandupTime, standupTime))
}

func (r *REST) addTimeTable(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}

	accessLevel, _ := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	logrus.Infof("Access level for %v in %v is %v", f.Get("user_id"), f.Get("channel_id"), accessLevel)
	if accessLevel > 3 {
		return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastPM)
	}

	usersText, weekdays, time, err := utils.SplitTimeTalbeCommand(ca.Text, r.conf.Translate.DaysDivider, r.conf.Translate.TimeDivider)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	users := strings.Split(usersText, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			c.String(http.StatusOK, r.conf.Translate.WrongUsernameError)
			continue
		}
		userID, userName := utils.SplitUser(u)

		m, err := r.db.FindChannelMemberByUserID(userID, f.Get("channel_id"))
		if err != nil {
			m, err = r.db.CreateChannelMember(model.ChannelMember{
				UserID:    userID,
				ChannelID: f.Get("channel_id"),
			})
			if err != nil {
				continue
			}
		}

		tt, err := r.db.SelectTimeTable(m.ID)
		if err != nil {
			logrus.Infof("Timetable for this standuper does not exist. Creating...")
			ttNew, err := r.db.CreateTimeTable(model.TimeTable{
				ChannelMemberID: m.ID,
			})
			ttNew = utils.PrepareTimeTable(ttNew, weekdays, time)
			ttNew, err = r.db.UpdateTimeTable(ttNew)
			if err != nil {
				c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.CanNotUpdateTimetable, userName, err))
				continue
			}
			logrus.Infof("Timetable created id:%v", ttNew.ID)
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.TimetableCreated, userID, ttNew.Show()))
			continue
		}
		tt = utils.PrepareTimeTable(tt, weekdays, time)
		tt, err = r.db.UpdateTimeTable(tt)
		if err != nil {
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.CanNotUpdateTimetable, userName, err))
			continue
		}
		logrus.Infof("Timetable updated id:%v", tt.ID)
		c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.TimetableUpdated, userID, tt.Show()))
	}
	return nil
}

func (r *REST) showTimeTable(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}

	users := strings.Split(ca.Text, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			c.String(http.StatusOK, r.conf.Translate.WrongUsernameError)
			continue
		}
		userID, userName := utils.SplitUser(u)

		m, err := r.db.FindChannelMemberByUserID(userID, f.Get("channel_id"))
		if err != nil {
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.NotAStanduper, userName))
			continue
		}
		tt, err := r.db.SelectTimeTable(m.ID)
		if err != nil {
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.NoTimetableSet, userName))
			continue
		}
		c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.TimetableShow, userName, tt.Show()))
	}
	return nil
}

func (r *REST) removeTimeTable(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}

	accessLevel, _ := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	logrus.Infof("Access level for %v in %v is %v", f.Get("user_id"), f.Get("channel_id"), accessLevel)
	if accessLevel > 3 {
		return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastPM)
	}

	users := strings.Split(ca.Text, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			c.String(http.StatusOK, r.conf.Translate.WrongUsernameError)
			continue
		}
		userID, userName := utils.SplitUser(u)

		m, err := r.db.FindChannelMemberByUserID(userID, f.Get("channel_id"))
		if err != nil {
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.NotAStanduper, userName))
			continue
		}
		tt, err := r.db.SelectTimeTable(m.ID)
		if err != nil {
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.NoTimetableSet, userName))
			continue
		}
		err = r.db.DeleteTimeTable(tt.ID)
		if err != nil {
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.CanNotDeleteTimetable, userName))
			continue
		}
		c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.TimetableDeleted, userName))
	}
	return nil
}

func (r *REST) reportByProject(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}

	accessLevel, _ := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	logrus.Infof("Access level for %v in %v is %v", f.Get("user_id"), f.Get("channel_id"), accessLevel)
	if accessLevel > 3 {
		return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastPM)
	}

	commandParams := strings.Fields(ca.Text)
	if len(commandParams) != 3 {
		return c.String(http.StatusOK, r.conf.Translate.WrongNArgs)
	}
	channelName := strings.Replace(commandParams[0], "#", "", -1)
	channelID, err := r.db.GetChannelID(channelName)
	if err != nil {
		logrus.Errorf("rest: GetChannelID failed: %v\n", err)
		return c.String(http.StatusOK, "Неверное название проекта!")
	}

	channel, err := r.db.SelectChannel(channelID)
	if err != nil {
		logrus.Errorf("rest: SelectChannel failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}

	dateFrom, err := time.Parse("2006-01-02", commandParams[1])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}
	dateTo, err := time.Parse("2006-01-02", commandParams[2])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}

	report, err := r.report.StandupReportByProject(channel, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("rest: StandupReportByProject: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}

	text := ""
	text += report.ReportHead
	if len(report.ReportBody) == 0 {
		text += r.conf.Translate.ReportNoData
		return c.String(http.StatusOK, text)
	}
	for _, t := range report.ReportBody {
		text += t.Text
		if r.conf.TeamMonitoringEnabled {
			cd, err := teammonitoring.GetCollectorData(r.conf, "projects", channel.ChannelName, t.Date.Format("2006-01-02"), t.Date.Format("2006-01-02"))
			if err != nil {
				continue
			}
			text += fmt.Sprintf(r.conf.Translate.ReportOnProjectCollectorData, cd.TotalCommits, utils.SecondsToHuman(cd.Worklogs))
		}
	}
	return c.String(http.StatusOK, text)
}

func (r *REST) reportByUser(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}

	commandParams := strings.Fields(ca.Text)
	if len(commandParams) != 3 {
		return c.String(http.StatusOK, r.conf.Translate.WrongNArgs)
	}
	username := strings.Replace(commandParams[0], "@", "", -1)
	user, err := r.db.SelectUserByUserName(username)
	if err != nil {
		return c.String(http.StatusOK, "User does not exist!")
	}

	accessLevel, _ := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	logrus.Infof("Access level for %v in %v is %v", f.Get("user_id"), f.Get("channel_id"), accessLevel)
	if f.Get("user_id") != user.UserID && accessLevel > 2 {
		return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastAdminOrOwner)
	}

	dateFrom, err := time.Parse("2006-01-02", commandParams[1])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}
	dateTo, err := time.Parse("2006-01-02", commandParams[2])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}

	report, err := r.report.StandupReportByUser(user.UserID, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("rest: StandupReportByUser failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}

	text := ""
	text += report.ReportHead
	if len(report.ReportBody) == 0 {
		text += r.conf.Translate.ReportNoData
		return c.String(http.StatusOK, text)
	}
	for _, t := range report.ReportBody {
		text += t.Text
		if r.conf.TeamMonitoringEnabled {
			cd, err := teammonitoring.GetCollectorData(r.conf, "users", user.UserID, t.Date.Format("2006-01-02"), t.Date.Format("2006-01-02"))
			if err != nil {
				continue
			}
			text += fmt.Sprintf(r.conf.Translate.ReportCollectorDataUser, cd.TotalCommits, utils.SecondsToHuman(cd.Worklogs))
		}
	}
	return c.String(http.StatusOK, text)
}

func (r *REST) reportByProjectAndUser(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}

	commandParams := strings.Fields(ca.Text)
	if len(commandParams) != 4 {
		return c.String(http.StatusOK, r.conf.Translate.WrongNArgs)
	}

	channelName := strings.Replace(commandParams[0], "#", "", -1)
	channelID, err := r.db.GetChannelID(channelName)
	if err != nil {
		logrus.Errorf("rest: GetChannelID failed: %v\n", err)
		return c.String(http.StatusOK, r.conf.Translate.WrongProjectName)
	}

	channel, err := r.db.SelectChannel(channelID)
	if err != nil {
		logrus.Errorf("rest: SelectChannel failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}

	username := strings.Replace(commandParams[1], "@", "", -1)

	user, err := r.db.SelectUserByUserName(username)
	if err != nil {
		return c.String(http.StatusOK, r.conf.Translate.NoSuchUserInWorkspace)
	}
	member, err := r.db.FindChannelMemberByUserName(user.UserName, channelID)
	if err != nil {
		return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.CanNotFindMember, user.UserID))
	}

	accessLevel, _ := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	logrus.Infof("Access level for %v in %v is %v", f.Get("user_id"), f.Get("channel_id"), accessLevel)
	if (f.Get("user_id") != member.UserID && f.Get("channel_id") != member.ChannelID) && accessLevel > 3 {
		return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastPMOrOwner)
	}

	dateFrom, err := time.Parse("2006-01-02", commandParams[2])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}
	dateTo, err := time.Parse("2006-01-02", commandParams[3])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}

	report, err := r.report.StandupReportByProjectAndUser(channel, member.UserID, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("rest: StandupReportByProjectAndUser failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}

	text := ""
	text += report.ReportHead
	if len(report.ReportBody) == 0 {
		text += r.conf.Translate.ReportNoData
		return c.String(http.StatusOK, text)
	}
	for _, t := range report.ReportBody {
		text += t.Text
		if r.conf.TeamMonitoringEnabled {
			data := fmt.Sprintf("%v/%v", member.UserID, channel.ChannelName)
			cd, err := teammonitoring.GetCollectorData(r.conf, "user-in-project", data, t.Date.Format("2006-01-02"), t.Date.Format("2006-01-02"))
			if err != nil {
				continue
			}
			text += fmt.Sprintf(r.conf.Translate.ReportCollectorDataUser, cd.TotalCommits, utils.SecondsToHuman(cd.Worklogs))
		}
	}
	return c.String(http.StatusOK, text)
}

func (r *REST) getAccessLevel(userID, channelID string) (int, error) {
	user, err := r.db.SelectUser(userID)
	if err != nil {
		return 0, err
	}
	if userID == r.conf.ManagerSlackUserID {
		return 1, nil
	}
	if user.IsAdmin() {
		return 2, nil
	}
	if r.db.UserIsPMForProject(userID, channelID) {
		return 3, nil
	}
	return 4, nil
}

func (r *REST) comedianIsInChannel(channelID string) bool {
	_, err := r.db.SelectChannel(channelID)
	if err != nil {
		logrus.Errorf("SelectChannel failed: %v", err)
		return false
	}
	return true
}

func (r *REST) validateRequest(c echo.Context, f url.Values) (FullSlackForm, error) {
	var ca FullSlackForm
	if !r.comedianIsInChannel(f.Get("channel_id")) {
		return ca, errors.New(r.conf.Translate.ComedianIsNotInChannel)
	}
	if err := r.decoder.Decode(&ca, f); err != nil {
		return ca, errors.New("I could not decode your command. Please, check if it is correct and try again")
	}
	if err := ca.Validate(); err != nil {
		return ca, err
	}
	return ca, nil
}

func (r *REST) processCommand(c echo.Context, f url.Values) ([]string, string, string, int, error) {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return []string{}, "", "", 0, err
	}

	accessLevel, err := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	if err != nil {
		logrus.Errorf("getAccessLevel Failed: %v", err)
		return []string{}, "", "", 0, err
	}
	parts := strings.Split(ca.Text, "/")

	if len(parts) > 1 {
		users := strings.Split(strings.TrimSpace(parts[0]), " ")
		return users, strings.TrimSpace(parts[1]), ca.ChannelID, accessLevel, nil
	}
	users := strings.Split(ca.Text, " ")
	return users, "developer", ca.ChannelID, accessLevel, nil
}
