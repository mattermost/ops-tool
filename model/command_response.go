package model

type Color struct {
	Color  string `yaml:"color"`
	Status string `yaml:"status"`
}

type CommandResponseType string

const (
	CommandResponseTypeEphemeral CommandResponseType = "ephemeral"
	CommandResponseTypeInChannel CommandResponseType = "in_channel"
	CommandResponseTypeDialog    CommandResponseType = "dialog"
)

type CommandResponse struct {
	Type    CommandResponseType
	Dialog  Dialog
	Message Message
}
