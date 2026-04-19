package dockercompose

import (
	"github.com/LucasAVasco/falcula/provider/base"
	"github.com/LucasAVasco/falcula/provider/dockercompose/info"
)

// ServiceOpts is the base options for a docker-compose service
type ServiceOpts = base.ServiceOpts

// Service represents a generic docker-compose service
type Service struct {
	*base.Service
	Info *info.DockerComposeInfo
}

func NewService(service *base.Service, info *info.DockerComposeInfo) *Service {
	return &Service{
		Service: service,
		Info:    info,
	}
}
