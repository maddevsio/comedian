package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
}

const (
	commandAddAdmin               = "/adminadd"
	commandRemoveAdmin            = "/adminremove"
	commandListAdmins             = "/adminlist"
	commandAddUser                = "/comedianadd"
	commandRemoveUser             = "/comedianremove"
	commandListUsers              = "/comedianlist"
	commandAddTime                = "/standuptimeset"
	commandRemoveTime             = "/standuptimeremove"
	commandListTime               = "/standuptime"
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

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	r := &REST{
		db:      conn,
		echo:    e,
		conf:    c,
		decoder: decoder,
		report:  rep,
		slack:   s,
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

func (r *REST) senderIsAdmin(form url.Values) bool {

	slackUserID := form.Get("user_id")
	channelID := form.Get("channel_id")
	userIsAdmin := r.db.IsAdmin(slackUserID, channelID)
	return (slackUserID == r.conf.ManagerSlackUserID) || (userIsAdmin == true)
}

func (r *REST) addUserCommand(c echo.Context, f url.Values) error {
	if !r.senderIsAdmin(f) {
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
		return c.String(http.StatusBadRequest, "Укажите пользователей, которых нужно добавить")
	}
	logrus.Infof("Users: %v", users)
	for _, u := range users {

		slackUserID, userName := splitUser(u)
		channelName, err := r.slack.GetChannelName(ca.ChannelID)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}
		user, err := r.db.FindStandupUserInChannelByUserID(slackUserID, ca.ChannelID)
		logrus.Errorf("Rest FindStandupUserInChannelByUserID failed: %v", err)
		if err != nil {
			_, err = r.db.CreateStandupUser(model.StandupUser{
				SlackUserID: slackUserID,
				SlackName:   userName,
				ChannelID:   ca.ChannelID,
				Channel:     channelName,
				Role:        "user",
			})
			if err != nil {
				logrus.Errorf("rest: CreateStandupUser failed: %v\n", err)
				c.String(http.StatusBadRequest, fmt.Sprintf("failed to create user :%v\n", err))
				continue
			}
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.AddUser, userName))
			continue
		}
		if user.SlackUserID == slackUserID && user.ChannelID == ca.ChannelID {
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.UserExist, slackUserID))
			continue
		}
		st, err := r.db.GetChannelStandupTime(ca.ChannelID)
		if err != nil {
			logrus.Errorf("rest: GetChannelStandupTime failed: %v\n", err)
		}
		if st.Time == int64(0) {
			c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.AddUserNoStandupTime, userName))
			continue
		}
	}
	return nil
}

func (r *REST) addAdminCommand(c echo.Context, f url.Values) error {
	if !r.senderIsAdmin(f) {
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
	result := strings.Split(ca.Text, "|")
	slackUserID := strings.Replace(result[0], "<@", "", -1)
	userName := strings.Replace(result[1], ">", "", -1)
	channelName, err := r.slack.GetChannelName(ca.ChannelID)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	user, err := r.db.FindStandupUserInChannelByUserID(slackUserID, ca.ChannelID)
	if err != nil {
		_, err := r.db.CreateStandupUser(model.StandupUser{
			SlackUserID: slackUserID,
			SlackName:   userName,
			ChannelID:   ca.ChannelID,
			Channel:     channelName,
			Role:        "admin",
		})
		if err != nil {
			logrus.Errorf("rest: CreateStandupUser failed: %v\n", err)
			return c.String(http.StatusBadRequest, fmt.Sprintf("failed to create user :%v\n", err))
		}
		message := fmt.Sprintf(r.conf.Translate.PMAssigned, ca.ChannelID, channelName)
		err = r.slack.SendUserMessage(slackUserID, message)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
	}
	if user.SlackName == userName && user.ChannelID == ca.ChannelID {
		return c.String(http.StatusOK, r.conf.Translate.UserExist)
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.AddAdmin, userName))
}

func (r *REST) removeUserCommand(c echo.Context, f url.Values) error {
	if !r.senderIsAdmin(f) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
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
		return c.String(http.StatusBadRequest, "Укажите пользователей, которых нужно удалить")
	}

	for _, u := range users {
		slackUserID, userName := splitUser(u)
		user, err := r.db.FindStandupUserInChannelByUserID(slackUserID, ca.ChannelID)
		if err != nil {
			logrus.Errorf("rest: FindStandupUserInChannelByUserID failed: %v\n", err)
			c.String(http.StatusOK, fmt.Sprintf("Пользователь <@%v> не стэндапит в этом канале\n", slackUserID))
			continue
		}
		err = r.db.DeleteStandupUser(user.SlackName, ca.ChannelID)
		if err != nil {
			logrus.Errorf("rest: DeleteStandupUser failed: %v\n", err)
			c.String(http.StatusBadRequest, fmt.Sprintf("failed to delete user :%v\n", err))
			continue
		}
		c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.DeleteUser, userName))
	}
	return nil
}

func (r *REST) removeAdminCommand(c echo.Context, f url.Values) error {
	if !r.senderIsAdmin(f) {
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

	result := strings.Split(ca.Text, "|")
	slackUserID := strings.Replace(result[0], "<@", "", -1)
	userName := strings.Replace(result[1], ">", "", -1)

	user, err := r.db.FindStandupUserInChannelByUserID(slackUserID, ca.ChannelID)
	if err != nil {
		logrus.Errorf("rest: DeleteStandupUser failed: %v\n", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to delete user :%v\n", err))
	}

	err = r.db.DeleteAdmin(userName, ca.ChannelID)
	if err != nil {
		logrus.Errorf("rest: DeleteStandupUser failed: %v\n", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to delete user :%v\n", err))
	}
	message := fmt.Sprintf(r.conf.Translate.PMRemoved, user.ChannelID, user.Channel)
	err = r.slack.SendUserMessage(slackUserID, message)
	if err != nil {
		logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.DeleteAdmin, userName))
}

func (r *REST) listUsersCommand(c echo.Context, f url.Values) error {
	logrus.Printf("%+v\n", f)
	var ca ChannelIDForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: listUsersCommand Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: listUsersCommand Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	users, err := r.db.ListStandupUsersByChannelID(ca.ChannelID)
	if err != nil {
		logrus.Errorf("rest: ListStandupUsersByChannelID: %v\n", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to list users :%v\n", err))
	}
	var userNames []string
	for _, user := range users {
		userNames = append(userNames, "<@"+user.SlackName+">")
	}
	if len(userNames) < 1 {
		return c.String(http.StatusOK, r.conf.Translate.ListNoStandupers)
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.ListStandupers, strings.Join(userNames, ", ")))
}

func (r *REST) listAdminsCommand(c echo.Context, f url.Values) error {
	if !r.senderIsAdmin(f) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	logrus.Printf("%+v\n", f)
	var ca ChannelIDForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("rest: listUsersCommand Decode failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("rest: listUsersCommand Validate failed: %v\n", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	users, err := r.db.ListAdminsByChannelID(ca.ChannelID)
	if err != nil {
		logrus.Errorf("rest: ListStandupUsersByChannelID: %v\n", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to list users :%v\n", err))
	}
	var userNames []string
	for _, user := range users {
		userNames = append(userNames, "<@"+user.SlackName+">")
	}
	if len(userNames) < 1 {
		return c.String(http.StatusOK, r.conf.Translate.ListNoAdmins)
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.ListAdmins, strings.Join(userNames, ", ")))
}

func (r *REST) addTime(c echo.Context, f url.Values) error {
	if !r.senderIsAdmin(f) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
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

	st, err := r.db.GetChannelStandupTime(ca.ChannelID)
	logrus.Infof("If channel standup time exis: %v, %v", st, err)
	if err != nil {
		logrus.Errorf("GetChannelStandupTime failed: %v", err)
		st, _ = r.db.CreateStandupTime(model.StandupTime{
			ChannelID: ca.ChannelID,
			Channel:   ca.ChannelName,
			Time:      timeInt,
		})
		return nil
	}
	st.Time = timeInt
	_, err = r.db.UpdateStandupTime(st)
	if err != nil {
		logrus.Errorf("rest: UpdateStandupTime failed: %v\n", err)
	}
	standupUsers, err := r.db.ListStandupUsersByChannelID(ca.ChannelID)
	if err != nil {
		logrus.Errorf("rest: ListStandupUsersByChannelID failed: %v\n", err)
	}
	if len(standupUsers) == 0 {
		return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.AddStandupTimeNoUsers, st.Time))
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.AddStandupTime, st.Time))
}

func (r *REST) removeTime(c echo.Context, f url.Values) error {
	if !r.senderIsAdmin(f) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
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
	st, err := r.db.ListStandupUsersByChannelID(ca.ChannelID)
	if len(st) != 0 {
		return c.String(http.StatusOK, r.conf.Translate.RemoveStandupTimeWithUsers)
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.RemoveStandupTime, ca.ChannelName))
}

func (r *REST) listTime(c echo.Context, f url.Values) error {
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
	if err != nil {
		logrus.Errorf("rest: GetChannelStandupTime failed: %v\n", err)
		if err.Error() == "sql: no rows in result set" {
			return c.String(http.StatusOK, r.conf.Translate.ShowNoStandupTime)
		} else {
			return c.String(http.StatusBadRequest, fmt.Sprintf("failed to list time :%v\n", err))
		}
	}
	return c.String(http.StatusOK, fmt.Sprintf(r.conf.Translate.ShowStandupTime, standupTime.Time))
}

///report_by_project #collector-test 2018-07-24 2018-07-26
func (r *REST) reportByProject(c echo.Context, f url.Values) error {
	if !r.senderIsAdmin(f) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	var ca ChannelIDTextForm
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

	logrus.Infof("channelName: %v", channelName)

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
	data, err := r.getCollectorData("projects", channelName, commandParams[1], commandParams[2])
	if err != nil {
		logrus.Errorf("rest: getCollectorData failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}
	channelID, err := r.db.GetChannelID(channelName)
	logrus.Infof("channelID: %v", channelID)
	if err != nil {
		logrus.Errorf("rest: GetChannelID failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}
	report, err := r.report.StandupReportByProject(channelID, dateFrom, dateTo, data)
	if err != nil {
		logrus.Errorf("rest: StandupReportByProject: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}
	return c.String(http.StatusOK, report)
}

///report_by_user @Anatoliy 2018-07-24 2018-07-26
func (r *REST) reportByUser(c echo.Context, f url.Values) error {
	if !r.senderIsAdmin(f) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	var ca FullSlackForm
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
	user, err := r.db.FindStandupUserByUserName(username)
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
	data, err := r.getCollectorData("users", user.SlackUserID, commandParams[1], commandParams[2])
	if err != nil {
		logrus.Errorf("rest: getCollectorData failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}
	report, err := r.report.StandupReportByUser(user.SlackUserID, dateFrom, dateTo, data)
	if err != nil {
		logrus.Errorf("rest: StandupReportByUser failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}
	return c.String(http.StatusOK, report)
}

///report_by_project_and_user #collector-test @Anatoliy 2018-07-24 2018-07-26
func (r *REST) reportByProjectAndUser(c echo.Context, f url.Values) error {
	if !r.senderIsAdmin(f) {
		return c.String(http.StatusOK, r.conf.Translate.AccessDenied)
	}
	var ca FullSlackForm
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
	channelID, channelName := splitChannel(commandParams[0])
	userID, _ := splitUser(commandParams[1])
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
	pu := channelName + "/" + userID
	data, err := r.getCollectorData("projects-users", pu, commandParams[2], commandParams[3])
	if err != nil {
		logrus.Errorf("rest: getCollectorData failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}

	user, err := r.db.FindStandupUserInChannelByUserID(userID, channelID)
	if err != nil {
		return c.String(http.StatusOK, r.conf.Translate.ReportByProjectAndUser)
	}
	report, err := r.report.StandupReportByProjectAndUser(channelID, user.SlackUserID, dateFrom, dateTo, data)
	if err != nil {
		logrus.Errorf("rest: StandupReportByProjectAndUser failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}
	return c.String(http.StatusOK, report)
}

func (r *REST) getCollectorData(getDataOn, data, dateFrom, dateTo string) ([]byte, error) {
	linkURL := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s", r.conf.CollectorURL, getDataOn, data, dateFrom, dateTo)
	logrus.Infof("rest: getCollectorData request URL: %s", linkURL)
	req, err := http.NewRequest("GET", linkURL, nil)
	if err != nil {
		logrus.Errorf("rest: http.NewRequest failed: %v\n", err)
		return nil, err
	}
	token := r.conf.CollectorToken
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Errorf("rest: http.DefaultClient.Do(req) failed: %v\n", err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("rest: ioutil.ReadAll(res.Body) failed: %v\n", err)
		return nil, err
	}
	logrus.Infof("rest: getCollectorData responce body: %s", string(body))
	return body, nil

}

func splitChannel(channel string) (string, string) {
	channelSeparate := strings.Split(channel, "|")
	channelID := strings.Replace(channelSeparate[0], "<#", "", -1)
	channelName := strings.Replace(channelSeparate[1], ">", "", -1)
	return channelID, channelName
}

func splitUser(user string) (string, string) {
	userFull := strings.Split(user, "|")
	userID := strings.Replace(userFull[0], "<@", "", -1)
	userName := strings.Replace(userFull[1], ">", "", -1)
	return userID, userName
}
