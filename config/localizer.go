package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

//Translate struct makes translation data
type Translate struct {
	UserExist            string
	AddUserNoStandupTime string
	AddUser              string
	AddAdmin             string
	PMExists             string
	PMAdded              string
	AccessDenied         string
	DeleteUser           string
	DeleteAdmin          string
	DeletePM             string
	ListNoStandupers     string
	ListNoAdmins         string
	ListStandupers       string
	ListAdmins           string
	ListNoPMs            string
	ListPMs              string
	UserIsNotPM          string

	AddStandupTimeNoUsers      string
	AddStandupTime             string
	RemoveStandupTimeWithUsers string
	RemoveStandupTime          string
	ShowNoStandupTime          string
	ShowStandupTime            string
	WrongNArgs                 string

	NoWorklogs          string
	NoCommits           string
	NoStandup           string
	HasWorklogs         string
	HasCommits          string
	HasStandup          string
	IsRook              string
	NotifyAllDone       string
	NotifyNotAll        string
	NotifyManagerNotAll string
	NotifyUsersWarning  string
	NotifyDirectMessage string

	ReportByProjectAndUser       string
	ReportOnProjectHead          string
	ReportOnProjectCollectorData string
	ReportOnUserHead             string
	ReportOnProjectAndUserHead   string
	ReportNoData                 string
	ReportDate                   string
	ReportStandupFromUser        string
	ReportIgnoredStandup         string
	ReportShowChannel            string
	ReportCollectorDataUser      string
	UserDidNotStandup            string
	UserDidStandup               string
	UserDidNotStandupInChannel   string
	UserDidStandupInChannel      string
	PMAssigned                   string
	PMRemoved                    string

	DateError1 string
	DateError2 string

	HelloManager    string
	StandupAccepted string

	WrongUsernameError string

	SelectUsersToAdd        string
	SelectUsersToDelete     string
	CanNotFindMember        string
	SelectUsersToAddAsAdmin string
	NoSuchUserInWorkspace   string
	UserNotAdmin            string
	WrongProjectName        string

	DaysDivider                 string
	TimeDivider                 string
	TimetableNoUsers            string
	TimetableCreated            string
	TimetableUpdated            string
	CanNotUpdateTimetable       string
	NotAStanduper               string
	NoTimetableSet              string
	TimetableShow               string
	CanNotDeleteTimetable       string
	TimetableDeleted            string
	IndividualStandupersWarning string
	IndividualStandupersLate    string
	EmptyTimetable              string

	TimetableShowMonday    string
	TimetableShowTuesday   string
	TimetableShowWednesday string
	TimetableShowThursday  string
	TimetableShowFriday    string
	TimetableShowSaturday  string
	TimetableShowSunday    string

	ComedianIsNotInChannel string

	StandupHandleUserNotAssigned          string
	StandupHandleOneDayOneStandup         string
	StandupHandleCouldNotSaveStandup      string
	StandupHandleNoProblemsMentioned      string
	StandupHandleNoYesterdayWorkMentioned string
	StandupHandleNoTodayPlansMentioned    string
	StandupHandleUpdatedStandup           string
	StandupHandleCreatedStandup           string

	ErrorRooksReportWeekend string
	ReportHeaderMonday      string
	ReportHeader            string

	AccessAtLeastPM           string
	AccessAtLeastAdmin        string
	AccessAtLeastSuperAdmin   string
	AccessAtLeastAdminOrOwner string
	AccessAtLeastPMOrOwner    string
}

// GetTranslation sets translation files for config
func GetTranslation(lang string) (Translate, error) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	wd, err := os.Getwd()
	if err != nil {
		return Translate{}, err
	}
	if !strings.HasSuffix(wd, "comedian") {
		wd = filepath.Dir(wd)
	}
	_, err = bundle.LoadMessageFile(fmt.Sprintf("%s/config/ru.toml", wd))
	if err != nil {
		return Translate{}, err
	}
	_, err = bundle.LoadMessageFile(fmt.Sprintf("%s/config/en.toml", wd))
	if err != nil {
		return Translate{}, err
	}
	localizer := i18n.NewLocalizer(bundle, lang)
	if err != nil {
		logrus.Errorf("slack: GetLocalizer failed: %v\n", err)
		return Translate{}, err
	}
	m := make(map[string]string)
	r := []string{
		"noWorklogs", "noCommits", "noStandup", "hasWorklogs",
		"hasCommits", "hasStandup", "isRook", "notifyAllDone",
		"notifyNotAll", "notifyManagerNotAll", "notifyUsersWarning",
		"notifyDirectMessage",
		"reportByProjectAndUser", "reportOnProjectHead", "reportOnProjectCollectorData", "reportOnUserHead",
		"reportOnProjectAndUserHead", "reportNoData", "reportDate",
		"reportStandupFromUser", "reportIgnoredStandup", "reportShowChannel",
		"reportCollectorDataUser",
		"helloManager", "standupAccepted",
		"userExist",
		"pmExists", "pmAdded",
		"addUserNoStandupTime",
		"addUser",
		"addAdmin",
		"accessDenied",
		"deleteUser",
		"deleteAdmin",
		"DeletePM",
		"listNoStandupers",
		"listNoAdmins",
		"listStandupers",
		"listAdmins",
		"listNoPMs",
		"listPMs",
		"addStandupTimeNoUsers",
		"addStandupTime",
		"removeStandupTimeWithUsers",
		"removeStandupTime",
		"showNoStandupTime",
		"showStandupTime",
		"wrongNArgs",
		"dateError1", "dateError2",
		"userDidNotStandup", "userDidStandup",
		"userDidNotStandupInChannel", "userDidStandupInChannel",
		"PMAssigned",
		"PMRemoved",
		"wrongUsernameError",

		"selectUsersToAdd",
		"selectUsersToDelete",
		"CanNotFindMember",
		"selectUsersToAddAsAdmin",
		"noSuchUserInWorkspace",
		"userNotAdmin",
		"wrongProjectName",
		"DaysDivider",
		"TimeDivider",
		"TimetableNoUsers",
		"TimetableCreated",
		"TimetableUpdated",
		"CanNotUpdateTimetable",
		"NotAStanduper",
		"NoTimetableSet",
		"TimetableShow",
		"CanNotDeleteTimetable",
		"TimetableDeleted",
		"IndividualStandupersWarning",
		"IndividualStandupersLate",
		"EmptyTimetable",
		"TimetableShowMonday",
		"TimetableShowTuesday",
		"TimetableShowWednesday",
		"TimetableShowThursday",
		"TimetableShowFriday",
		"TimetableShowSaturday",
		"TimetableShowSunday",
		"ComedianIsNotInChannel",

		"StandupHandleUserNotAssigned",
		"StandupHandleOneDayOneStandup",
		"StandupHandleCouldNotSaveStandup",
		"StandupHandleNoProblemsMentioned",
		"StandupHandleNoYesterdayWorkMentioned",
		"StandupHandleNoTodayPlansMentioned",
		"StandupHandleUpdatedStandup",
		"StandupHandleCreatedStandup",
		"ErrorRooksReportWeekend",
		"ReportHeaderMonday",
		"ReportHeader",
		"UserIsNotPM",
		"AccessAtLeastPM",
		"AccessAtLeastAdmin",
		"AccessAtLeastSuperAdmin",
		"AccessAtLeastAdminOrOwner",
		"AccessAtLeastPMOrOwner",
	}

	for _, t := range r {
		translated, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: t})
		if err != nil {
			logrus.Errorf("slack: Localize failed: %v\n", err)
			return Translate{}, err
		}
		m[t] = translated
	}

	t := Translate{
		UserExist:            m["userExist"],
		AddUserNoStandupTime: m["addUserNoStandupTime"],
		AddUser:              m["addUser"],
		AddAdmin:             m["addAdmin"],
		AccessDenied:         m["accessDenied"],
		PMExists:             m["pmExists"],
		PMAdded:              m["pmAdded"],

		DeleteUser:       m["deleteUser"],
		DeleteAdmin:      m["deleteAdmin"],
		DeletePM:         m["DeletePM"],
		ListNoStandupers: m["listNoStandupers"],
		ListNoAdmins:     m["listNoAdmins"],
		ListStandupers:   m["listStandupers"],
		ListAdmins:       m["listAdmins"],
		ListNoPMs:        m["listNoPMs"],
		ListPMs:          m["listPMs"],

		AddStandupTimeNoUsers:        m["addStandupTimeNoUsers"],
		AddStandupTime:               m["addStandupTime"],
		RemoveStandupTimeWithUsers:   m["removeStandupTimeWithUsers"],
		RemoveStandupTime:            m["removeStandupTime"],
		ShowNoStandupTime:            m["showNoStandupTime"],
		ShowStandupTime:              m["showStandupTime"],
		WrongNArgs:                   m["wrongNArgs"],
		NoWorklogs:                   m["noWorklogs"],
		NoCommits:                    m["noCommits"],
		NoStandup:                    m["noStandup"],
		HasWorklogs:                  m["hasWorklogs"],
		HasCommits:                   m["hasCommits"],
		HasStandup:                   m["hasStandup"],
		IsRook:                       m["isRook"],
		NotifyAllDone:                m["notifyAllDone"],
		NotifyNotAll:                 m["notifyNotAll"],
		NotifyManagerNotAll:          m["notifyManagerNotAll"],
		NotifyUsersWarning:           m["notifyUsersWarning"],
		NotifyDirectMessage:          m["notifyDirectMessage"],
		ReportByProjectAndUser:       m["reportByProjectAndUser"],
		ReportOnProjectHead:          m["reportOnProjectHead"],
		ReportOnProjectCollectorData: m["reportOnProjectCollectorData"],
		ReportOnUserHead:             m["reportOnUserHead"],
		ReportOnProjectAndUserHead:   m["reportOnProjectAndUserHead"],
		ReportNoData:                 m["reportNoData"],
		ReportDate:                   m["reportDate"],
		ReportStandupFromUser:        m["reportStandupFromUser"],
		ReportIgnoredStandup:         m["reportIgnoredStandup"],
		ReportShowChannel:            m["reportShowChannel"],
		ReportCollectorDataUser:      m["reportCollectorDataUser"],
		DateError1:                   m["dateError1"],
		DateError2:                   m["dateError2"],
		HelloManager:                 m["helloManager"],
		StandupAccepted:              m["standupAccepted"],
		UserDidNotStandup:            m["userDidNotStandup"],
		UserDidStandup:               m["userDidStandup"],
		UserDidNotStandupInChannel:   m["userDidNotStandupInChannel"],
		UserDidStandupInChannel:      m["userDidStandupInChannel"],
		PMAssigned:                   m["PMAssigned"],
		PMRemoved:                    m["PMRemoved"],

		WrongUsernameError: m["wrongUsernameError"],

		SelectUsersToAdd:        m["selectUsersToAdd"],
		SelectUsersToDelete:     m["selectUsersToDelete"],
		CanNotFindMember:        m["CanNotFindMember"],
		SelectUsersToAddAsAdmin: m["selectUsersToAddAsAdmin"],
		NoSuchUserInWorkspace:   m["noSuchUserInWorkspace"],
		UserNotAdmin:            m["userNotAdmin"],
		WrongProjectName:        m["wrongProjectName"],

		DaysDivider:                 m["DaysDivider"],
		TimeDivider:                 m["TimeDivider"],
		TimetableNoUsers:            m["TimetableNoUsers"],
		TimetableCreated:            m["TimetableCreated"],
		TimetableUpdated:            m["TimetableUpdated"],
		CanNotUpdateTimetable:       m["CanNotUpdateTimetable"],
		NotAStanduper:               m["NotAStanduper"],
		NoTimetableSet:              m["NoTimetableSet"],
		TimetableShow:               m["TimetableShow"],
		CanNotDeleteTimetable:       m["CanNotDeleteTimetable"],
		TimetableDeleted:            m["TimetableDeleted"],
		IndividualStandupersWarning: m["IndividualStandupersWarning"],
		IndividualStandupersLate:    m["IndividualStandupersLate"],
		EmptyTimetable:              m["EmptyTimetable"],
		TimetableShowMonday:         m["TimetableShowMonday"],
		TimetableShowTuesday:        m["TimetableShowTuesday"],
		TimetableShowWednesday:      m["TimetableShowWednesday"],
		TimetableShowThursday:       m["TimetableShowThursday"],
		TimetableShowFriday:         m["TimetableShowFriday"],
		TimetableShowSaturday:       m["TimetableShowSaturday"],
		TimetableShowSunday:         m["TimetableShowSunday"],
		ComedianIsNotInChannel:      m["ComedianIsNotInChannel"],

		StandupHandleUserNotAssigned:          m["StandupHandleUserNotAssigned"],
		StandupHandleOneDayOneStandup:         m["StandupHandleOneDayOneStandup"],
		StandupHandleCouldNotSaveStandup:      m["StandupHandleCouldNotSaveStandup"],
		StandupHandleNoProblemsMentioned:      m["StandupHandleNoProblemsMentioned"],
		StandupHandleNoYesterdayWorkMentioned: m["StandupHandleNoYesterdayWorkMentioned"],
		StandupHandleNoTodayPlansMentioned:    m["StandupHandleNoTodayPlansMentioned"],
		StandupHandleUpdatedStandup:           m["StandupHandleUpdatedStandup"],
		StandupHandleCreatedStandup:           m["StandupHandleCreatedStandup"],

		ErrorRooksReportWeekend: m["ErrorRooksReportWeekend"],
		ReportHeaderMonday:      m["ReportHeaderMonday"],
		ReportHeader:            m["ReportHeader"],
		UserIsNotPM:             m["UserIsNotPM"],

		AccessAtLeastPM:           m["AccessAtLeastPM"],
		AccessAtLeastAdmin:        m["AccessAtLeastAdmin"],
		AccessAtLeastSuperAdmin:   m["AccessAtLeastSuperAdmin"],
		AccessAtLeastAdminOrOwner: m["AccessAtLeastAdminOrOwner"],
		AccessAtLeastPMOrOwner:    m["AccessAtLeastPMOrOwner"],
	}

	return t, nil
}
