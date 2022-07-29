package server

import (
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name        string
		inputCmd    string
		wantRootCmd string
		wantCmdText string
		wantArgs    map[string]string
		wantErr     bool
	}{
		{
			name:        "empty",
			inputCmd:    "",
			wantRootCmd: "",
			wantCmdText: "",
			wantArgs:    map[string]string{},
			wantErr:     false,
		},
		{
			name:        "command without args",
			inputCmd:    "/command",
			wantRootCmd: "command",
			wantCmdText: "",
			wantArgs:    map[string]string{},
			wantErr:     false,
		},
		{
			name:        "long command without args",
			inputCmd:    "/command with other stuff",
			wantRootCmd: "command",
			wantCmdText: "with other stuff",
			wantArgs:    map[string]string{},
			wantErr:     false,
		},
		{
			name:        "command with args",
			inputCmd:    "/command --arg1=value1 --arg2=value2",
			wantRootCmd: "command",
			wantCmdText: "",
			wantArgs: map[string]string{
				"arg1": "value1",
				"arg2": "value2",
			},
			wantErr: false,
		},
		{
			name:        "command with args without value (should return true)",
			inputCmd:    "/command --arg1 --arg2=value2",
			wantRootCmd: "command",
			wantArgs: map[string]string{
				"arg1": "true",
				"arg2": "value2",
			},
			wantErr: false,
		},
		{
			name:        "command with args containing spaces",
			inputCmd:    `/command --arg1="value 1"`,
			wantRootCmd: "command",
			wantArgs: map[string]string{
				"arg1": "value 1",
			},
			wantErr: false,
		},
		{
			name:        "long command with all types of args",
			inputCmd:    `/command with other stuff --arg1="value 1" --arg2=value2 --arg3`,
			wantRootCmd: "command",
			wantCmdText: "with other stuff",
			wantArgs: map[string]string{
				"arg1": "value 1",
				"arg2": "value2",
				"arg3": "true",
			},
			wantErr: false,
		},
		{
			name:     "returns error on badly formed command",
			inputCmd: `/command with other stuff --arg1="`,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRootCmd, gotCmdText, gotArgs, err := ParseCommand(tt.inputCmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRootCmd != tt.wantRootCmd {
				t.Errorf("ParseCommand() gotRootCmd = %v, want %v", gotRootCmd, tt.wantRootCmd)
			}
			if gotCmdText != tt.wantCmdText {
				t.Errorf("ParseCommand() gotCmdText = %v, want %v", gotCmdText, tt.wantCmdText)
			}
			if len(tt.wantArgs) > 0 && !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("ParseCommand() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
