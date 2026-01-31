// Package base implements a base provider.
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

func New(multi *multiplexer.Multiplexer, name string) *Provider {
	return &Provider{
		Multiplexer: multi,
		Name:        name,
		Color:       colorgen.Next(),
	}
}

func (p *Provider) GetName() string {
	return p.Name
}
