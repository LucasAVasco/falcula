// Package process implements a service and provider for system process services
package process

import (
	"github.com/LucasAVasco/falcula/provider/base"
)

// Command command to run
type Command struct {
	Shell   bool     // true to run the command in a shell
	Command []string // Command and its arguments
}

// ProviderOpts is the options for a process provider
type ProviderOpts = base.ProviderOpts

// ProviderConfig is the configuration for a process provider
type ProviderConfig = base.ProviderConfig

// Provider is a provider to generate process services
type Provider struct {
	*base.Provider
	prepareCmd *Command
	mainCmd    *Command
}

// New creates a new process provider. Generates the steps from the commands. If the command is nil, the step will do nothing (you can
// disable a step by setting the command to nil)
func New(config *ProviderConfig, prepareCmd *Command, mainCmd *Command) *Provider {
	return &Provider{
		Provider:   base.NewProvider(config),
		prepareCmd: prepareCmd,
		mainCmd:    mainCmd,
	}
}

func (p *Provider) NewService(opts *base.ServiceOpts) *Service {
	return &Service{
		Service:    p.Provider.NewService("", opts),
		prepareCmd: p.prepareCmd,
		mainCmd:    p.mainCmd,
	}
}
