package chat

import "github.com/nlopes/slack"

// Chat inteface should be implemented for all messengers(facebook, slack, telegram, whatever)
type Chat interface {
	Run()
	SendMessage(string, string) error
	SendUserMessage(string, string) error
	SendReportMessage(string, string, []slack.Attachment) error
}
