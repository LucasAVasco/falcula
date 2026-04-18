// Package dockercompose implements a service and provider for docker-compose
package dockercompose

import (
	"github.com/LucasAVasco/falcula/provider/base"
	"github.com/LucasAVasco/falcula/provider/dockercompose/info"
	"github.com/LucasAVasco/falcula/service/iface"
)

// Provider is a docker-compose service provider (generate docker-compose services)
type Provider struct {
	*base.Provider
	info *info.DockerComposeInfo
}

func New(config *base.ProviderConfig, composeFile string) *Provider {
	return &Provider{
		Provider: base.NewProvider(config),
		info:     info.NewComposeInfo(composeFile),
	}
}

func (p *Provider) NewService(name string, opts *base.ServiceOpts) *Service {
	return NewService(p.Provider.NewService(name, opts), p.info)
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
		Service: p.NewService("build", &opts.ServiceOpts),
		opts:    opts,
	}
}

func (p *Provider) NewUpService(platform string, opts *base.ServiceOpts) iface.Service {
	return &UpService{
		Service:  p.NewService("up", opts),
		platform: platform,
	}
}

func (p *Provider) NewDownService(opts *DownServiceOpts) iface.Service {
	if opts == nil {
		opts = &DownServiceOpts{}
	}
	return &DownService{
		Service: p.NewService("down", &opts.ServiceOpts),
		opts:    opts,
	}
}

// NewPushService returns a new push service. Opts must not be nil
func (p *Provider) NewPushService(opts *PushServiceOpts) iface.Service {
	return &PushService{
		Service: p.NewService("push", &opts.ServiceOpts),
		opts:    opts,
	}
}
