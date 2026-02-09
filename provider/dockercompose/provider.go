// Package dockercompose implements a service and provider for docker-compose
package dockercompose

import (
	"github.com/LucasAVasco/falcula/multiplexer"
	"github.com/LucasAVasco/falcula/provider/base"
	"github.com/LucasAVasco/falcula/provider/dockercompose/info"
	"github.com/LucasAVasco/falcula/service/iface"
)

// Provider is a docker-compose service provider (generate docker-compose services)
type Provider struct {
	*base.Provider
	info *info.DockerComposeInfo
}

func New(multi *multiplexer.Multiplexer, name string, composeFile string) *Provider {
	return &Provider{
		Provider: base.NewProvider(multi, name),
		info:     info.NewComposeInfo(composeFile),
	}
}

func (p *Provider) NewService(name string) *Service {
	return NewService(p.Provider.NewService(name), p.info)
}

func (p *Provider) AddDefaultPushImage(image string) {
	p.info.AddDefaultPushImage(image)
}

func (p *Provider) AddDefaultPushImages(images []string) {
	p.info.AddDefaultPushImages(images)
}

// NewBuildService returns a new build service. If opts is nil, default values will be used. If the opts.Builds is nil, a default value that
// builds all buildable services for the current platform will be used
func (p *Provider) NewBuildService(opts *BuildServiceOpts) *BuildService {
	// Default values for the options
	if opts == nil {
		opts = &BuildServiceOpts{}
	}

	if opts.Builds == nil {
		opts.Builds = []*BuildInfo{
			{
				ServicesNames: []string{}, // All buildable services
				Platforms:     []string{}, // Current platform
			},
		}
	}

	// Build service
	return &BuildService{
		Service: p.NewService("build"),
		opts:    opts,
	}
}

func (p *Provider) NewUpService(platform string) iface.Service {
	return &UpService{
		Service:  p.NewService("up"),
		provider: p,
		platform: platform,
	}
}

func (p *Provider) NewDownService() iface.Service {
	return &DownService{
		Service: p.NewService("down"),
	}
}

// NewPushService returns a new push service. Opts must not be nil
func (p *Provider) NewPushService(opts *PushServiceOpts) iface.Service {
	return &PushService{
		Service: p.NewService("push"),
		opts:    opts,
	}
}
