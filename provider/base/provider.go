// Package base implements a base provider and service.
package base

import (
	"github.com/LucasAVasco/falcula/colorgen"
	"github.com/LucasAVasco/falcula/multiplexer"
	"github.com/LucasAVasco/falcula/service/iface"

	"github.com/fatih/color"
)

// ProviderConfig is the configuration for a provider. It is required to create a provider
type ProviderConfig struct {
	Multiplexer        *multiplexer.Multiplexer
	Name               string
	DefaultServiceOpts iface.Opts // The default options for a service. If not provided, a default value will be used
}

// Provider represents a provider. Provides the basic data required by a provider. Any provider should inherit from this
type Provider struct {
	Color  *color.Color
	Config *ProviderConfig
}

func NewProvider(config *ProviderConfig) *Provider {
	return &Provider{
		Color:  colorgen.Next(),
		Config: config,
	}
}

func (p *Provider) GetName() string {
	return p.Config.Name
}

// NewService creates a new service. Its name is automatically prefixed with the provider name. If empty, the provider name is used
func (p *Provider) NewService(name string, opts *ServiceOpts) *Service {
	if name == "" {
		name = p.Config.Name
	} else {
		name = p.Config.Name + "." + name
	}

	config := ServiceConfig{
		Multiplexer: p.Config.Multiplexer,
		Color:       p.Color,
		Name:        name,
		Opts:        p.Config.DefaultServiceOpts,
	}

	return NewService(&config, opts)
}
