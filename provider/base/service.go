package base

import (
	"github.com/LucasAVasco/falcula/multiplexer"
	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/service/iface"
	"github.com/fatih/color"
)

// ServiceConfig is a structure that holds the configurations for a service. It is required to create a service
type ServiceConfig struct {
	Multiplexer *multiplexer.Multiplexer
	Color       *color.Color
	Name        string
	Opts        iface.Opts
}

// ServiceOpts is a structure that holds the options for a service. It is optional
type ServiceOpts struct {
	StartDisabled *bool `lua:"start_disabled"`
}

// Service represents a base service with the basic data required by a service. Any service should inherit from this
type Service struct {
	Config ServiceConfig
}

// NewService creates a new service. The config is required and the opts is optional
func NewService(config *ServiceConfig, opts *ServiceOpts) *Service {
	s := Service{
		Config: *config,
	}

	if opts != nil {
		if opts.StartDisabled != nil {
			s.Config.Opts.StartDisabled = *opts.StartDisabled
		}
	}

	return &s
}

func (s *Service) GetName() string {
	return s.Config.Name
}

func (s *Service) GetOpts() *iface.Opts {
	return &s.Config.Opts
}

// NewProcessOptions returns a new process.Options struct configured for the service. It will use the provider multiplexer and color.
//
// It will also set `Wait` to `false` and `ManualStart` to `true`. If you use the process adapters at '../adapter/', you do not need to
// manually start the process because they automatically starts the process if it is not already started
func (s *Service) NewProcessOptions() *process.Options {
	return &process.Options{
		Multiplexer: s.Config.Multiplexer,
		Name:        s.Config.Name,
		Wait:        false,
		ManualStart: true,
		Color:       s.Config.Color,
	}
}
