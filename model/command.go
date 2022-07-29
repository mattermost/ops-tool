package model

type Command struct {
	Command        string
	Name           string
	Description    string
	Usage          string
	CommandHandler func(mmCommand *MMSlashCommand, args map[string]string) (*CommandResponse, error)
	DialogHandler  func(submission *DialogSubmission, args map[string]string) (*CommandResponse, error)
}
