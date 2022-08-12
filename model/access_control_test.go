package model

import "testing"

func TestAccessControl_IsEmpty(t *testing.T) {
	type fields struct {
		TeamID      []string
		TeamName    []string
		ChannelID   []string
		ChannelName []string
		UserID      []string
		UserName    []string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "empty",
			fields: fields{},
			want:   true,
		},
		{
			name: "full",
			fields: fields{
				TeamID:      []string{"1"},
				TeamName:    []string{"2"},
				ChannelID:   []string{"3"},
				ChannelName: []string{"4"},
				UserID:      []string{"5"},
				UserName:    []string{"6"},
			},
			want: false,
		},
		{
			name: "team id not empty",
			fields: fields{
				TeamID: []string{"1"},
			},
			want: false,
		},
		{
			name: "team name not empty",
			fields: fields{
				TeamName: []string{"2"},
			},
			want: false,
		},
		{
			name: "channel id not empty",
			fields: fields{
				ChannelID: []string{"3"},
			},
			want: false,
		},
		{
			name: "channel name not empty",
			fields: fields{
				ChannelName: []string{"4"},
			},
			want: false,
		},
		{
			name: "user id not empty",
			fields: fields{
				UserID: []string{"5"},
			},
			want: false,
		},
		{
			name: "user name not empty",
			fields: fields{
				UserName: []string{"6"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AccessControl{
				TeamID:      tt.fields.TeamID,
				TeamName:    tt.fields.TeamName,
				ChannelID:   tt.fields.ChannelID,
				ChannelName: tt.fields.ChannelName,
				UserID:      tt.fields.UserID,
				UserName:    tt.fields.UserName,
			}
			if got := a.IsEmpty(); got != tt.want {
				t.Errorf("AccessControl.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
