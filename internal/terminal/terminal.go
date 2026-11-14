// Package terminal detects whether falcula runs in an interactive terminal (where the text user interface can be shown)
package terminal

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// ciVariables are environment variables set by continuous integration providers. Any non-empty value means a CI environment. `CI` is
// handled separately because some users set it to "false"
var ciVariables = []string{
	"CONTINUOUS_INTEGRATION",
	"GITHUB_ACTIONS",
	"GITLAB_CI",
	"BUILDKITE",
	"CIRCLECI",
	"TRAVIS",
	"JENKINS_URL",
	"TEAMCITY_VERSION",
	"TF_BUILD",
	"CODEBUILD_BUILD_ID",
	"DRONE",
	"BITBUCKET_BUILD_NUMBER",
}

// Reason is why the text user interface can not be shown. The empty reason means it can be shown
type Reason string

const (
	ReasonNone             Reason = ""
	ReasonCI               Reason = "running in a CI environment"
	ReasonInputNotTerminal Reason = "the standard input is not a terminal"
	ReasonOutputRedirected Reason = "the standard output is not a terminal"
)

// environment is what the detection depends on. Replaced in tests
type environment struct {
	getenv     func(key string) string
	isTerminal func(file *os.File) bool
	stdin      *os.File
	stdout     *os.File
}

func isTerminal(file *os.File) bool {
	return file != nil && term.IsTerminal(int(file.Fd()))
}

// NonInteractiveReason returns why the text user interface can not be shown (CI environment, redirected output or no terminal to read the
// user input from). Returns `ReasonNone` if it can be shown
func NonInteractiveReason() Reason {
	return environment{
		getenv:     os.Getenv,
		isTerminal: isTerminal,
		stdin:      os.Stdin,
		stdout:     os.Stdout,
	}.nonInteractiveReason()
}

func (e environment) nonInteractiveReason() Reason {
	switch {
	case e.isCI():
		return ReasonCI
	case !e.isTerminal(e.stdout):
		return ReasonOutputRedirected
	case !e.isTerminal(e.stdin):
		return ReasonInputNotTerminal
	default:
		return ReasonNone
	}
}

func (e environment) isCI() bool {
	switch strings.ToLower(strings.TrimSpace(e.getenv("CI"))) {
	case "", "0", "false", "no", "off":
	default:
		return true
	}

	for _, key := range ciVariables {
		if e.getenv(key) != "" {
			return true
		}
	}

	return false
}
