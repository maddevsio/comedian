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
	commandAddAdmin               = "/admin_add"
	commandRemoveAdmin            = "/admin_remove"
	commandListAdmins             = "/admin_list"
	commandAddPM                  = "/pm_add"
	commandRemovePM               = "/pm_remove"
	commandListPMs                = "/pm_list"
	commandAddUser                = "/comedian_add"
	commandRemoveUser             = "/comedian_remove"
	commandListUsers              = "/comedian_list"
	commandAddTime                = "/standup_time_set"
	commandRemoveTime             = "/standup_time_remove"
	commandListTime               = "/standup_time"
	commandAddTimeTable           = "/timetable_set"
	commandRemoveTimeTable        = "/timetable_remove"
	commandShowTimeTable          = "/timetable_show"
	commandReportByProject        = "/report_by_project"
	commandReportByUser           = "/report_by_user"
	commandReportByProjectAndUser = "/report_by_project_and_user"
)

// NewRESTAPI creates API for Slack commands
func NewRESTAPI(slack *chat.Slack) (*REST, error) {
	e := echo.New()
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	rep, err := reporting.NewReporter(slack.Conf)
	if err != nil {
		logrus.Errorf("rest: NewReporter failed: %v\n", err)
		return nil, err
	}

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
	logrus.Infof("Rest context: %v", c)
	form, err := c.FormParams()
	logrus.Infof("Rest form: %v", form)
	if err != nil {
		logrus.Errorf("rest: c.FormParams failed: %v\n", err)
	}
	if command := form.Get("command"); command != "" {
		switch command {
		case commandAddUser:
			return r.addUserCommand(c, form)
		case commandListUsers:
			return r.listUsersCommand(c, form)
		case commandRemoveUser:
			return r.removeUserCommand(c, form)
		case commandAddAdmin:
			return r.addAdminCommand(c, form)
		case commandRemoveAdmin:
			return r.removeAdminCommand(c, form)
		case commandListAdmins:
			return r.listAdminsCommand(c, form)
		case commandAddPM:
			return r.addPMCommand(c, form)
		case commandRemovePM:
			return r.removePMCommand(c, form)
		case commandListPMs:
			return r.listPMsCommand(c, form)
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
		case commandReportByProjectAndUser:
			return r.reportByProjectAndUser(c, form)
		default:
			return c.String(http.StatusNotImplemented, "Not implemented")
		}
	}
	return c.JSON(http.StatusMethodNotAllowed, "Command not allowed")
}

func (r *REST) addUserCommand(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}

	accessLevel, _ := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	if accessLevel > 3 {
		return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastPM)
	}
	users := strings.Split(ca.Text, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			return c.String(http.StatusOK, r.conf.Translate.WrongUsernameError)
		}
		userID, userName := utils.SplitUser(u)
		user, err := r.db.FindChannelMemberByUserID(userID, ca.ChannelID)
		if err != nil {
			logrus.Errorf("Rest FindChannelMemberByUserID failed: %v", err)
			chanMember, _ := r.db.CreateChannelMember(model.ChannelMember{
				UserID:    userID,
				ChannelID: ca.ChannelID,
			})
			logrus.Infof("ChannelMember created! ID:%v", chanMember.ID)
		}

		if user.UserID == userID && user.ChannelID == ca.ChannelID {
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.UserExist, userID))
			continue
		}

		st, err := r.db.GetChannelStandupTime(ca.ChannelID)
		if err != nil {
			logrus.Errorf("rest: GetChannelStandupTime failed: %v\n", err)
		}
		logrus.Infof("channel standup time: %v", st)

		if st == int64(0) {
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.AddUserNoStandupTime, userName))
			continue
		}

		c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.AddUser, userName))
	}
	return nil
}

func (r *REST) removeUserCommand(c echo.Context, f url.Values) error {
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
	if len(users) < 1 {
		return c.String(http.StatusBadRequest, r.conf.Translate.SelectUsersToDelete)
	}
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			return c.String(http.StatusOK, r.conf.Translate.WrongUsernameError)
		}
		userID, userName := utils.SplitUser(u)
		user, err := r.db.FindChannelMemberByUserID(userID, ca.ChannelID)
		if err != nil {
			logrus.Errorf("rest: FindChannelMemberByUserID failed: %v\n", err)
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.CanNotFindMember, userID))
			continue
		}
		err = r.db.DeleteChannelMember(user.UserID, ca.ChannelID)
		if err != nil {
			logrus.Errorf("rest: DeleteChannelMember failed: %v\n", err)
			c.String(http.StatusBadRequest, fmt.Sprintf("failed to delete user :%v\n", err))
			continue
		}
		c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.DeleteUser, userName))
	}
	return nil
}

func (r *REST) listUsersCommand(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	users, err := r.db.ListChannelMembers(ca.ChannelID)
	if err != nil {
		logrus.Errorf("rest: ListChannelMembers: %v\n", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to list users :%v\n", err))
	}
	var userIDs []string
	for _, user := range users {
		userIDs = append(userIDs, "<@"+user.UserID+">")
	}
	if len(userIDs) < 1 {
		return c.String(http.StatusOK, r.conf.Translate.ListNoStandupers)
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.ListStandupers, strings.Join(userIDs, ", ")))
}

func (r *REST) addPMCommand(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	accessLevel, _ := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	logrus.Infof("Access level for %v in %v is %v", f.Get("user_id"), f.Get("channel_id"), accessLevel)
	if accessLevel > 2 {
		return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastAdmin)
	}
	users := strings.Split(ca.Text, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			return c.String(http.StatusOK, r.conf.Translate.WrongUsernameError)
		}
		userID, _ := utils.SplitUser(u)
		isAdmin := r.db.UserIsPMForProject(userID, ca.ChannelID)
		if !isAdmin {
			_, err := r.db.CreatePM(model.ChannelMember{
				UserID:    userID,
				ChannelID: ca.ChannelID,
			})
			if err != nil {
				logrus.Errorf("rest: CreatePM failed: %v\n", err)
				c.String(http.StatusOK, fmt.Sprintf("failed to create user :%v\n", err))
			}
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.PMAdded, userID))
			continue
		}
		c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.PMExists, userID))
	}
	return nil
}

func (r *REST) removePMCommand(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	accessLevel, _ := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	logrus.Infof("Access level for %v in %v is %v", f.Get("user_id"), f.Get("channel_id"), accessLevel)
	if accessLevel > 2 {
		return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastAdmin)
	}
	users := strings.Split(ca.Text, " ")
	if len(users) < 1 {
		return c.String(http.StatusBadRequest, r.conf.Translate.SelectUsersToDelete)
	}
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			return c.String(http.StatusOK, r.conf.Translate.WrongUsernameError)
		}
		userID, userName := utils.SplitUser(u)
		user, err := r.db.FindChannelMemberByUserID(userID, ca.ChannelID)
		if err != nil {
			logrus.Errorf("rest: FindChannelMemberByUserID failed: %v\n", err)
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.CanNotFindMember, userID))
			continue
		}
		if user.RoleInChannel != "PM" {
			logrus.Errorf("rest: User %v is not PM in %v! Skip\n", userName, user.ChannelID)
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.UserIsNotPM, userID))
			continue
		}
		err = r.db.DeleteChannelMember(user.UserID, ca.ChannelID)
		if err != nil {
			logrus.Errorf("rest: DeleteChannelMember failed: %v\n", err)
			c.String(http.StatusBadRequest, fmt.Sprintf("failed to delete user :%v\n", err))
			continue
		}
		c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.DeletePM, userName))
	}
	return nil
}

func (r *REST) listPMsCommand(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	users, err := r.db.ListPMs(ca.ChannelID)
	if err != nil {
		logrus.Errorf("rest: ListPMs: %v\n", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to list users :%v\n", err))
	}
	var userIDs []string
	for _, user := range users {
		userIDs = append(userIDs, "<@"+user.UserID+">")
	}
	if len(userIDs) < 1 {
		return c.String(http.StatusOK, r.conf.Translate.ListNoPMs)
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.ListPMs, strings.Join(userIDs, ", ")))
}

func (r *REST) addAdminCommand(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	accessLevel, _ := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	logrus.Infof("Access level for %v in %v is %v", f.Get("user_id"), f.Get("channel_id"), accessLevel)
	if accessLevel > 1 {
		return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastSuperAdmin)
	}
	users := strings.Split(ca.Text, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			return c.String(http.StatusOK, r.conf.Translate.WrongUsernameError)
		}
		userID, userName := utils.SplitUser(u)
		user, err := r.db.SelectUser(userID)
		if err != nil {
			c.String(http.StatusOK, r.conf.Translate.NoSuchUserInWorkspace)
			continue
		}
		if user.Role == "admin" {
			c.String(http.StatusOK, "User is already admin!")
			continue
		}
		user.Role = "admin"
		_, err = r.db.UpdateUser(user)
		if err != nil {
			logrus.Errorf("rest: UpdateUser failed: %v\n", err)
		}
		message := r.conf.Translate.PMAssigned
		err = r.slack.SendUserMessage(userID, message)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.AddAdmin, userName))
	}

	return nil
}

func (r *REST) removeAdminCommand(c echo.Context, f url.Values) error {
	ca, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	accessLevel, _ := r.getAccessLevel(f.Get("user_id"), f.Get("channel_id"))
	logrus.Infof("Access level for %v in %v is %v", f.Get("user_id"), f.Get("channel_id"), accessLevel)
	if accessLevel > 1 {
		return c.String(http.StatusOK, r.conf.Translate.AccessAtLeastSuperAdmin)
	}
	users := strings.Split(ca.Text, " ")
	if len(users) < 1 {
		return c.String(http.StatusBadRequest, r.conf.Translate.SelectUsersToDelete)
	}

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			return c.String(http.StatusOK, r.conf.Translate.WrongUsernameError)
		}
		userID, userName := utils.SplitUser(u)
		user, err := r.db.SelectUser(userID)
		if err != nil {
			c.String(http.StatusOK, r.conf.Translate.NoSuchUserInWorkspace)
			continue
		}
		if user.Role != "admin" {
			c.String(http.StatusOK, r.conf.Translate.UserNotAdmin)
			continue
		}
		user.Role = ""
		_, err = r.db.UpdateUser(user)
		if err != nil {
			logrus.Errorf("rest: UpdateUser failed: %v\n", err)
		}
		message := fmt.Sprintf(r.conf.Translate.PMRemoved)
		err = r.slack.SendUserMessage(userID, message)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.DeleteAdmin, userName))
	}
	return nil
}

func (r *REST) listAdminsCommand(c echo.Context, f url.Values) error {
	_, err := r.validateRequest(c, f)
	if err != nil {
		logrus.Errorf("Validate Request Failed: %v", err)
		return c.String(http.StatusOK, err.Error())
	}

	admins, err := r.db.ListAdmins()
	if err != nil {
		logrus.Errorf("rest: ListChannelMembers: %v\n", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to list users :%v\n", err))
	}
	var userNames []string
	for _, admin := range admins {
		userNames = append(userNames, "<@"+admin.UserName+">")
	}
	if len(userNames) < 1 {
		return c.String(http.StatusOK, r.conf.Translate.ListNoAdmins)
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.ListAdmins, strings.Join(userNames, ", ")))
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
		return c.String(http.StatusOK, "Unexpected error occured when I tried to complete operation. Please, try again!")
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
		return c.String(http.StatusOK, "Unexpected error occured when I tried to complete operation. Please, try again!")
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
	if len(users) < 1 {
		return c.String(http.StatusOK, r.conf.Translate.TimetableNoUsers)
	}
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			c.String(http.StatusOK, r.conf.Translate.WrongUsernameError)
			continue
		}
		userID, userName := utils.SplitUser(u)

		m, err := r.db.FindChannelMemberByUserID(userID, f.Get("channel_id"))
		if err != nil {
			logrus.Errorf("FindChannelMemberByUserID failed: %v", err)
			m, err = r.db.CreateChannelMember(model.ChannelMember{
				UserID:    userID,
				ChannelID: f.Get("channel_id"),
			})
			if err != nil {
				logrus.Errorf("rest: CreateChannelMember failed: %v\n", err)
				c.String(http.StatusBadRequest, fmt.Sprintf("failed to create user:%v\n", err))
				continue
			}
		}

		tt, err := r.db.SelectTimeTable(m.ID)
		if err != nil {
			logrus.Infof("Timetable for this standuper does not exist. Creating...")
			ttNew, err := r.db.CreateTimeTable(model.TimeTable{
				ChannelMemberID: m.ID,
			})
			ttNew, err = r.prepareTimeTable(ttNew, weekdays, time)
			if err != nil {
				c.String(http.StatusOK, fmt.Sprintf("Could not fetch data into timetable for user <@%v>\n", userName))
				continue
			}
			ttNew, err = r.db.UpdateTimeTable(ttNew)
			if err != nil {
				c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.CanNotUpdateTimetable, userName, err))
				continue
			}
			logrus.Infof("Timetable created id:%v", ttNew.ID)
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.TimetableCreated, userID, ttNew.Show()))
			continue
		}
		tt, err = r.prepareTimeTable(tt, weekdays, time)
		if err != nil {
			c.String(http.StatusOK, fmt.Sprintf("Could not fetch data into timetable for user <@%v>\n", userName))
			continue
		}
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
	if len(users) < 1 {
		return c.String(http.StatusBadRequest, "Select standupers to show their timetables")
	}

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
	if len(users) < 1 {
		return c.String(http.StatusBadRequest, "Select standupers to delete their timetables")
	}

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
		return c.String(http.StatusOK, r.conf.Translate.UserExist)
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

func (r *REST) prepareTimeTable(tt model.TimeTable, weekdays string, timeInt int64) (model.TimeTable, error) {
	if strings.Contains(weekdays, "mon") || strings.Contains(weekdays, "пн") {
		tt.Monday = timeInt
	}
	if strings.Contains(weekdays, "tue") || strings.Contains(weekdays, "вт") {
		tt.Tuesday = timeInt
	}
	if strings.Contains(weekdays, "wed") || strings.Contains(weekdays, "ср") {
		tt.Wednesday = timeInt
	}
	if strings.Contains(weekdays, "thu") || strings.Contains(weekdays, "чт") {
		tt.Thursday = timeInt
	}
	if strings.Contains(weekdays, "fri") || strings.Contains(weekdays, "пт") {
		tt.Friday = timeInt
	}
	if strings.Contains(weekdays, "sat") || strings.Contains(weekdays, "сб") {
		tt.Saturday = timeInt
	}
	if strings.Contains(weekdays, "sun") || strings.Contains(weekdays, "вс") {
		tt.Sunday = timeInt
	}
	return tt, nil
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
