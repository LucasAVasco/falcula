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

func (p *Provider) NewBuildService(onlyBuild bool) *BuildService {
	return &BuildService{
		Service:   p.NewService("build"),
		onlyBuild: onlyBuild,
	}
}

func (p *Provider) NewUpService() iface.Service {
	return &UpService{
		Service:  p.NewService("up"),
		provider: p,
	}
}

func (p *Provider) NewDownService() iface.Service {
	return &DownService{
		Service:  p.NewService("down"),
		provider: p,
	}
}

func (p *Provider) NewPushService(repositories []string) iface.Service {
	return &PushService{
		Service:      p.NewService("push"),
		repositories: repositories,
	}
}
