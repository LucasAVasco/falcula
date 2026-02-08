// Package base implements a base provider and service.
package base

import (
	"github.com/LucasAVasco/falcula/colorgen"
	"github.com/LucasAVasco/falcula/multiplexer"

	"github.com/fatih/color"
)

// Provider represents a provider. Provides the basic data required by a provider. Any provider should inherit from this
type Provider struct {
	Multiplexer *multiplexer.Multiplexer
	Name        string
	Color       *color.Color
}

func NewProvider(multi *multiplexer.Multiplexer, name string) *Provider {
	return &Provider{
		Multiplexer: multi,
		Name:        name,
		Color:       colorgen.Next(),
	}
}

func (p *Provider) GetName() string {
	return p.Name
}

// NewService creates a new service. Its name is automatically prefixed with the provider name. If empty, the provider name is used
func (p *Provider) NewService(name string) *Service {
	if name == "" {
		name = p.Name
	} else {
		name = p.Name + "." + name
	}

	return &Service{
		Provider: p,
		Name:     name,
	}
}
