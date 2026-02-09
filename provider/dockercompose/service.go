package dockercompose

import (
	"github.com/LucasAVasco/falcula/provider/base"
	"github.com/LucasAVasco/falcula/provider/dockercompose/info"
)

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
