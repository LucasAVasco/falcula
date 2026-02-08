// Package dockercompose implements a service and provider for docker-compose
package dockercompose

import (
	"path/filepath"

	"github.com/LucasAVasco/falcula/multiplexer"
	"github.com/LucasAVasco/falcula/provider/base"
	"github.com/LucasAVasco/falcula/service/iface"
)

// Provider is a docker-compose service provider (generate docker-compose services)
type Provider struct {
	*base.Provider
	composeFile string
	images      []string
}

func New(multi *multiplexer.Multiplexer, name string, composeFile string) *Provider {
	return &Provider{
		Provider:    base.NewProvider(multi, name),
		composeFile: filepath.Clean(composeFile),
		images:      []string{},
	}
}

func (p *Provider) AddImage(image string) {
	p.images = append(p.images, image)
}

func (p *Provider) AddImages(images []string) {
	p.images = append(p.images, images...)
}

func (p *Provider) NewBuildService(onlyBuild bool) *BuildService {
	return &BuildService{
		Service:   p.NewService("build"),
		provider:  p,
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
		provider:     p,
		images:       p.images,
		repositories: repositories,
	}
}
