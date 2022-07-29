package server

import (
	"strings"

	"github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
)

// ParseCommand parses a command from a string and returns the commands + the args
func ParseCommand(inputCmd string) (rootCmd string, cmdText string, args map[string]string, err error) {
	args = map[string]string{}

	res, err := shellwords.Parse(strings.TrimPrefix(inputCmd, "/"))
	if err != nil {
		return "", "", args, errors.Wrap(err, "parse slash command")
	}
	if len(res) == 0 {
		return "", "", args, nil
	}

	rootCmd = res[0]

	parsingArgs := false
	for _, x := range res[1:] {
		if !parsingArgs {
			if strings.HasPrefix(x, "--") {
				parsingArgs = true
			} else {
				cmdText += x + " "
			}
		}

		if parsingArgs {
			if !strings.HasPrefix(x, "--") {
				// ignore string that does not start with --
				continue
			}

			parts := strings.Split(x, "=")
			name := strings.TrimSpace(strings.TrimPrefix(parts[0], "--"))
			if len(parts) == 1 {
				args[name] = "true"
			} else {
				args[name] = strings.TrimSpace(parts[1])
			}
		}
	}

	cmdText = strings.TrimSpace(cmdText)
	return rootCmd, cmdText, args, nil
}
