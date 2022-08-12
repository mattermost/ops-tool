package config

import (
	"context"
	"os"
	"reflect"
	"testing"
)

func Test_parseConfigTemplate(t *testing.T) {
	tests := []struct {
		name      string
		injectEnv map[string]string
		input     []byte
		output    []byte
	}{
		{
			name:   "empty",
			input:  []byte(``),
			output: []byte(``),
		},
		{
			name: "environment variable must be present",
			injectEnv: map[string]string{
				"FOO": "bar",
			},
			input:  []byte(`{{ .Env.FOO }}`),
			output: []byte(`bar`),
		},
		{
			name: "input similar to our config without variables",
			input: []byte(`
listen_address: "0.0.0.0:8080"
commands:
  - name: jops
    token: "AbCdEfGhIjKlMnOpQrStUvWxYz"
	access_control:
	  user_name: ["user1", "user2"]`),
			output: []byte(`
listen_address: "0.0.0.0:8080"
commands:
  - name: jops
    token: "AbCdEfGhIjKlMnOpQrStUvWxYz"
	access_control:
	  user_name: ["user1", "user2"]`),
		},
		{
			name: "input similar to our config with variable",
			injectEnv: map[string]string{
				"MM_OPTOOLS_LISTEN_ADDRESS":  "0.0.0.0:8080",
				"MM_CMD_JOPS_TOKEN":          "AbCdEfGhIjKlMnOpQrStUvWxYz",
				"MM_CMD_JOPS_ACL_USER_NAMES": `["user1", "user2"]`,
			},
			input: []byte(`
listen_address: "{{ .Env.MM_OPTOOLS_LISTEN_ADDRESS }}"
commands:
  - name: jops
    token: "{{ .Env.MM_CMD_JOPS_TOKEN }}"
	access_control:
	  user_name: {{ .Env.MM_CMD_JOPS_ACL_USER_NAMES }}`),
			output: []byte(`
listen_address: "0.0.0.0:8080"
commands:
  - name: jops
    token: "AbCdEfGhIjKlMnOpQrStUvWxYz"
	access_control:
	  user_name: ["user1", "user2"]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for varName := range tt.injectEnv {
				os.Setenv(varName, tt.injectEnv[varName])
			}

			if got := parseConfigTemplate(context.TODO(), tt.input); !reflect.DeepEqual(got, tt.output) {
				if len(got) == 0 && len(tt.output) == 0 {
					// valid
					return
				}
				t.Errorf("parseConfigTemplate() = %v, want %v", got, tt.output)
			}
		})
	}
}
