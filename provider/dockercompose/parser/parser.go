// Package parser implements a parser for docker-compose files. This does not implements the full docker-compose specification, only the
// required ones for the 'docker-compose' provider
package parser

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

// Build represents the 'build' section of a docker-compose file
type Build struct {
	Dockerfile *string `yaml:"dockerfile"`
	Context    *string `yaml:"context"`
}

// Service represents a 'service' in a docker-compose file
type Service struct {
	Image *string `yaml:"image"`
	Build *Build  `yaml:"build"`
}

// File represents a docker-compose file
type File struct {
	Services map[string]Service `yaml:"services"`
}

// ParseFile parses a docker-compose file
func ParseFile(path string) (*File, error) {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var file File
	err = yaml.Unmarshal(fileContent, &file)
	if err != nil {
		return nil, fmt.Errorf("error parsing file: %w", err)
	}

	return &file, err
}
