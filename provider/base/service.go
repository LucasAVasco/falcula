package base

import "github.com/LucasAVasco/falcula/process"

// Service represents a base service with the basic data required by a service. Any service should inherit from this
type Service struct {
	Provider *Provider
	Name     string
}

func NewService(provider *Provider, name string) *Service {
	return &Service{
		Provider: provider,
		Name:     provider.Name + "." + name,
	}
}

func (s *Service) GetName() string {
	return s.Name
}

// NewProcessOptions returns a new process.Options struct configured for the service. It will use the provider multiplexer and color.
//
// It will also set `Wait` to `false` and `ManualStart` to `true`. If you use the process adapters at '../adapter/', you do not need to
// manually start the process because they automatically starts the process if it is not already started
func (s *Service) NewProcessOptions() *process.Options {
	return &process.Options{
		Multiplexer: s.Provider.Multiplexer,
		Name:        s.Name,
		Wait:        false,
		ManualStart: true,
		Color:       s.Provider.Color,
	}
}
