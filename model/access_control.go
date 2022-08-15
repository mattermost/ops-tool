package model

type AccessControl struct {
	TeamID      []string `yaml:"team_id"`
	TeamName    []string `yaml:"team_name"`
	ChannelID   []string `yaml:"channel_id"`
	ChannelName []string `yaml:"channel_name"`
	UserID      []string `yaml:"user_id"`
	UserName    []string `yaml:"user_name"`
}

// return true if all rules are empty
func (a *AccessControl) IsEmpty() bool {
	return len(a.TeamID) == 0 && len(a.TeamName) == 0 && len(a.ChannelID) == 0 && len(a.ChannelName) == 0 && len(a.UserID) == 0 && len(a.UserName) == 0
}
