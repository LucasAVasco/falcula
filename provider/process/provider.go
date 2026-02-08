// Package process implements a service and provider for system process services
package process

import (
	"github.com/LucasAVasco/falcula/multiplexer"
	"github.com/LucasAVasco/falcula/provider/base"
)

// Command command to run
type Command struct {
	Shell   bool     // true to run the command in a shell
	Command []string // Command and its arguments
}

type Provider struct {
	*base.Provider
	prepareCmd *Command
	mainCmd    *Command
}

// New creates a new process provider. Generates the steps from the commands. If the command is nil, the step will do nothing (you can
// disable a step by setting the command to nil)
func New(multi *multiplexer.Multiplexer, name string, prepareCmd *Command, mainCmd *Command) *Provider {
	return &Provider{
		Provider:   base.NewProvider(multi, name),
		prepareCmd: prepareCmd,
		mainCmd:    mainCmd,
	}
}

func (p *Provider) GetName() string {
	return p.Name
}

func (p *Provider) NewService() *Service {
	return &Service{
		provider:   p,
		name:       p.Name,
		prepareCmd: p.prepareCmd,
		mainCmd:    p.mainCmd,
	}
}
