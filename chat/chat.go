package chat

type (
	// Chat inteface should be implemented for all messengers(facebook, slack, telegram, whatever)
	Chat interface {
		Run() error
		SendMessage(string, string) error
	}
)
