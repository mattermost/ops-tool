package slashcommand

import (
	"testing"

	"github.com/mattermost/ops-tool/model"
)

func TestSlashCommand_accessControl(t *testing.T) {
	mmSlashCommand := &model.MMSlashCommand{
		TeamID:      "1",
		TeamName:    "Team 1",
		ChannelID:   "12345",
		ChannelName: "Channel 12345",
		UserID:      "54321",
		Username:    "User 54321",
	}

	tests := []struct {
		name    string
		acsCtrl model.AccessControl
		wantErr bool
	}{
		{
			name:    "empty access control must be allowed",
			acsCtrl: model.AccessControl{},
			wantErr: false,
		},
		{
			name: "valid team id",
			acsCtrl: model.AccessControl{
				TeamID: []string{"1"},
			},
			wantErr: false,
		},
		{
			name: "invalid team id",
			acsCtrl: model.AccessControl{
				TeamID: []string{"2"},
			},
			wantErr: true,
		},
		{
			name: "valid team name",
			acsCtrl: model.AccessControl{
				TeamName: []string{"Team 1"},
			},
			wantErr: false,
		},
		{
			name: "invalid team name",
			acsCtrl: model.AccessControl{
				TeamName: []string{"Team 2"},
			},
			wantErr: true,
		},
		{
			name: "valid channel id",
			acsCtrl: model.AccessControl{
				ChannelID: []string{"12345"},
			},
			wantErr: false,
		},
		{
			name: "invalid channel id",
			acsCtrl: model.AccessControl{
				ChannelID: []string{"12346"},
			},
			wantErr: true,
		},
		{
			name: "valid channel name",
			acsCtrl: model.AccessControl{
				ChannelName: []string{"Channel 12345"},
			},
			wantErr: false,
		},
		{
			name: "invalid channel name",
			acsCtrl: model.AccessControl{
				ChannelName: []string{"Channel 12346"},
			},
			wantErr: true,
		},
		{
			name: "valid user id",
			acsCtrl: model.AccessControl{
				UserID: []string{"54321"},
			},
			wantErr: false,
		},
		{
			name: "invalid user id",
			acsCtrl: model.AccessControl{
				UserID: []string{"543210"},
			},
			wantErr: true,
		},
		{
			name: "valid user name",
			acsCtrl: model.AccessControl{
				UserName: []string{"User 54321"},
			},
			wantErr: false,
		},
		{
			name: "invalid user name",
			acsCtrl: model.AccessControl{
				UserName: []string{"User 543210"},
			},
			wantErr: true,
		},
		{
			name: "valid all ids",
			acsCtrl: model.AccessControl{
				TeamID:    []string{"1"},
				ChannelID: []string{"12345"},
				UserID:    []string{"54321"},
			},
			wantErr: false,
		},
		{
			name: "one invalid id",
			acsCtrl: model.AccessControl{
				TeamID:    []string{"1"},
				ChannelID: []string{"wrong"},
				UserID:    []string{"54321"},
			},
			wantErr: true,
		},
		{
			name: "valid all names",
			acsCtrl: model.AccessControl{
				TeamName:    []string{"Team 1"},
				ChannelName: []string{"Channel 12345"},
				UserName:    []string{"User 54321"},
			},
			wantErr: false,
		},
		{
			name: "one invalid name",
			acsCtrl: model.AccessControl{
				TeamName:    []string{"Team 1"},
				ChannelName: []string{"Channel 12345"},
				UserName:    []string{"User wrong"},
			},
			wantErr: true,
		},
		{
			name: "all valid",
			acsCtrl: model.AccessControl{
				TeamID:      []string{"1"},
				ChannelID:   []string{"12345"},
				UserID:      []string{"54321"},
				TeamName:    []string{"Team 1"},
				ChannelName: []string{"Channel 12345"},
				UserName:    []string{"User 54321"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SlashCommand{
				AccessControl: tt.acsCtrl,
			}
			if err := s.accessControl(mmSlashCommand); (err != nil) != tt.wantErr {
				t.Errorf("SlashCommand.accessControl() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
