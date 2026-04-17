package project

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

// Command is a command to be executed. It can be either a single string (executed by the shell) or a list of strings specifying the command
// to be executed and its arguments (will not be executed by the shell)
type Command struct {
	String string
	List   []string
}

func (s *Command) UnmarshalYAML(data []byte) error {
	var single string

	// Try parsing as a single string first
	if err := yaml.Unmarshal(data, &single); err == nil {
		s.String = single
		return nil
	}

	// Fallback to parsing as a slice of strings
	var list []string
	if err := yaml.Unmarshal(data, &list); err != nil {
		return fmt.Errorf("error parsing command: %w", err)
	}
	s.List = list

	return nil
}

// IsNotEmpty returns true if the command has something to execute (either a shell command or a executable with arguments)
func (s *Command) IsNotEmpty() bool {
	if s.String == "" && len(s.List) == 0 {
		return false
	}
	return true
}
