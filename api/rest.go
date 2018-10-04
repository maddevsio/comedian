package api

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
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
	commandAddPM                  = "/pm_add"
	commandRemoveAdmin            = "/admin_remove"
	commandListAdmins             = "/admin_list"
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
func NewRESTAPI(c config.Config) (*REST, error) {
	e := echo.New()
	conn, err := storage.NewMySQL(c)
	if err != nil {
		logrus.Errorf("rest: NewMySQL failed: %v\n", err)
		return nil, err
	}
	rep, err := reporting.NewReporter(c)
	if err != nil {
		logrus.Errorf("rest: NewReporter failed: %v\n", err)
		return nil, err
	}

	s, err := chat.NewSlack(c)
	if err != nil {
		logrus.Errorf("rest: NewSlack failed: %v\n", err)
		return nil, err
	}

	api := slack.New(c.SlackToken)

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	r := &REST{
		db:      conn,
		echo:    e,
		conf:    c,
		decoder: decoder,
		report:  rep,
		slack:   s,
		api:     api,
	}

	r.initEndpoints()
	return r, nil
}

func (r *REST) initEndpoints() {
	r.echo.POST("/commands", r.handleCommands)
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
		case commandAddAdmin:
			return r.addAdminCommand(c, form)
		case commandAddPM:
			return r.addPMCommand(c, form)
		case commandRemoveUser:
			return r.removeUserCommand(c, form)
		case commandRemoveAdmin:
			return r.removeAdminCommand(c, form)
		case commandListUsers:
			return r.listUsersCommand(c, form)
		case commandListAdmins:
			return r.listAdminsCommand(c, form)
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
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	if !r.ComedianIsInChannel(f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.ComedianIsNotInChannel)
	}
	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: addUserCommand Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: addUserCommand Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	users := strings.Split(ca.Text, " ")
	if len(users) < 1 {
		return c.String(http.StatusBadRequest, r.conf.Translate.SelectUsersToAdd)
	}
	logrus.Infof("Users: %v", users)
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			return c.String(http.StatusOK, r.conf.Translate.WrongUsernameError)
		}
		userID, userName := utils.SplitUser(u)
		user, err := r.db.FindChannelMemberByUserID(userID, ca.ChannelID)
		if err != nil {
			logrus.Errorf("Rest FindChannelMemberByUserID failed: %v", err)
			_, err = r.db.CreateChannelMember(model.ChannelMember{
				UserID:    userID,
				ChannelID: ca.ChannelID,
			})
			if err != nil {
				logrus.Errorf("rest: CreateChannelMember failed: %v\n", err)
				c.String(http.StatusBadRequest, fmt.Sprintf("failed to create user :%v\n", err))
				continue
			}
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

func (r *REST) addPMCommand(c echo.Context, f url.Values) error {
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	if !r.ComedianIsInChannel(f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.ComedianIsNotInChannel)
	}
	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: addUserCommand Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: addUserCommand Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	users := strings.Split(ca.Text, " ")
	if len(users) < 1 {
		return c.String(http.StatusBadRequest, r.conf.Translate.SelectUsersToAdd)
	}
	logrus.Infof("Users: %v", users)
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

func (r *REST) removeUserCommand(c echo.Context, f url.Values) error {
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	if !r.ComedianIsInChannel(f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.ComedianIsNotInChannel)
	}
	var ca ChannelIDTextForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: removeUserCommand Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: removeUserCommand Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
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
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.UserDoesNotStandup, userID))
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
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	if !r.ComedianIsInChannel(f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.ComedianIsNotInChannel)
	}
	var ca ChannelIDForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: listUsersCommand Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: listUsersCommand Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
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

func (r *REST) addAdminCommand(c echo.Context, f url.Values) error {
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}

	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: addUserCommand Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: addUserCommand Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	users := strings.Split(ca.Text, " ")

	if len(users) < 1 {
		return c.String(http.StatusBadRequest, r.conf.Translate.SelectUsersToAddAsAdmin)
	}
	logrus.Infof("Users: %v", users)
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
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	var ca ChannelIDTextForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: removeAdminCommand Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: removeAdminCommand Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
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
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
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
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	if !r.ComedianIsInChannel(f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.ComedianIsNotInChannel)
	}
	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: addTime Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: addTime Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	result := strings.Split(ca.Text, ":")
	hours, err := strconv.Atoi(result[0])
	if err != nil {
		logrus.Errorf("rest: strconv.Atoi failed: %v\n", err)
		return err
	}
	munites, err := strconv.Atoi(result[1])
	if err != nil {
		logrus.Errorf("rest: strconv.Atoi failed: %v\n", err)
		return err
	}
	currentTime := time.Now()
	timeInt := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), hours, munites, 0, 0, time.Local).Unix()

	err = r.db.CreateStandupTime(timeInt, ca.ChannelID)
	if err != nil {
		logrus.Errorf("rest: CreateStandupTime failed: %v\n", err)
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
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	if !r.ComedianIsInChannel(f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.ComedianIsNotInChannel)
	}
	var ca ChannelForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: removeTime Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: removeTime Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	err := r.db.DeleteStandupTime(ca.ChannelID)
	if err != nil {
		logrus.Errorf("rest: DeleteStandupTime failed: %v\n", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to delete standup time :%v\n", err))
	}
	st, err := r.db.ListChannelMembers(ca.ChannelID)
	if len(st) != 0 {
		return c.String(http.StatusOK, r.conf.Translate.RemoveStandupTimeWithUsers)
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.RemoveStandupTime, ca.ChannelName))
}

func (r *REST) listTime(c echo.Context, f url.Values) error {
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	if !r.ComedianIsInChannel(f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.ComedianIsNotInChannel)
	}
	var ca ChannelIDForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: listTime Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: listTime Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	standupTime, err := r.db.GetChannelStandupTime(ca.ChannelID)
	logrus.Errorf("GetChannelStandupTime failed: %v", err)
	if err != nil || standupTime == int64(0) {
		return c.String(http.StatusOK, r.conf.Translate.ShowNoStandupTime)
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.ShowStandupTime, standupTime))
}

func (r *REST) addTimeTable(c echo.Context, f url.Values) error {
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	if !r.ComedianIsInChannel(f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.ComedianIsNotInChannel)
	}
	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: addTime Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: addTime Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	usersText, weekdays, time, err := utils.SplitTimeTalbeCommand(ca.Text, r.conf.Translate.DaysDivider, r.conf.Translate.TimeDivider)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	users := strings.Split(usersText, " ")
	if len(users) < 1 {
		return c.String(http.StatusBadRequest, r.conf.Translate.TimetableNoUsers)
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
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	if !r.ComedianIsInChannel(f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.ComedianIsNotInChannel)
	}
	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: listTime Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: listTime Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
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
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	if !r.ComedianIsInChannel(f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.ComedianIsNotInChannel)
	}
	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: removeTime Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: removeTime Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
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

///report_by_project #collector-test 2018-07-24 2018-07-26
func (r *REST) reportByProject(c echo.Context, f url.Values) error {
	var ca ChannelIDTextForm
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}

	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: reportByProject Decode failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: reportByProject Validate failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
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
			text += fmt.Sprintf(r.conf.Translate.ReportOnProjectCollectorData, cd.TotalCommits)
		}
	}
	return c.String(http.StatusOK, text)
}

///report_by_user @Anatoliy 2018-07-24 2018-07-26
func (r *REST) reportByUser(c echo.Context, f url.Values) error {
	var ca FullSlackForm
	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: reportByUser Decode failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: reportByUser Validate failed: %v\n", err)
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

///report_by_project_and_user #collector-test @Anatoliy 2018-07-24 2018-07-26
func (r *REST) reportByProjectAndUser(c echo.Context, f url.Values) error {
	var ca FullSlackForm

	if !r.userHasAccess(f.Get("user_id"), f.Get("channel_id")) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: reportByProjectAndUser Decode failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: reportByProjectAndUser Validate failed: %v\n", err)
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
		return c.String(http.StatusOK, r.conf.Translate.UserDoesNotStandup)
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

func (r *REST) userHasAccess(userID, channelID string) bool {
	user, err := r.db.SelectUser(userID)
	if err != nil {
		logrus.Error(err)
		return false
	}
	if (userID == r.conf.ManagerSlackUserID) || user.IsAdmin() {
		logrus.Infof("User: %s has admin access!", user.UserName)
		return true
	}
	if r.db.UserIsPMForProject(userID, channelID) {
		logrus.Infof("User: %s is Project PM", user.UserName)
		return true
	}
	logrus.Infof("User: %s does not have admin access!", user.UserName)
	return false
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

func (r *REST) ComedianIsInChannel(channelID string) bool {
	_, err := r.db.SelectChannel(channelID)
	if err != nil {
		logrus.Errorf("SelectChannel failed: %v", err)
		return false
	}
	return true
}
