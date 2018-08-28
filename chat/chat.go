package chat

// Chat inteface should be implemented for all messengers(facebook, slack, telegram, whatever)
type Chat interface {
	Run() error
	SendMessage(string, string) error
	SendUserMessage(string, string) error
}
