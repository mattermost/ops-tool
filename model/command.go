package model

import "context"

type Command struct {
	Command        string
	Name           string
	Description    string
	Usage          string
	Plugin         string
	CommandHandler func(ctx context.Context, mmCommand *MMSlashCommand, args map[string]string) (*CommandResponse, error)
	DialogHandler  func(ctx context.Context, submission *DialogSubmission, args map[string]string) (*CommandResponse, error)
}
