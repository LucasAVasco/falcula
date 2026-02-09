// Package info provides information about a docker-compose file
package info

import (
	"fmt"
	"time"

	"github.com/LucasAVasco/falcula/provider/dockercompose/parser"
)

// DockerComposeInfo holds information about a docker-compose file
type DockerComposeInfo struct {
	composeFilePath      string
	composeFile          *parser.File
	composeFileReadTime  time.Time
	defaultBuildServices []string
	defaultPushImages    []string
}

// NewComposeInfo creates a new DockerComposeInfo for the provided docker-compose file
func NewComposeInfo(composeFilePath string) *DockerComposeInfo {
	return &DockerComposeInfo{
		composeFilePath:      composeFilePath,
		defaultBuildServices: []string{},
		defaultPushImages:    []string{},
	}
}

// GetComposeFilePath returns the path to the docker-compose file. Same value as passed to NewComposeInfo without modification
func (p *DockerComposeInfo) GetComposeFilePath() string {
	return p.composeFilePath
}

// GetComposeFile returns the parsed docker-compose file. You must not modify this. It will automatically refresh the file every 5 seconds
// and return the latest version
func (p *DockerComposeInfo) GetComposeFile() (*parser.File, error) {
	if p.composeFile == nil || time.Since(p.composeFileReadTime) > 5*time.Second {
		var err error
		p.composeFile, err = parser.ParseFile(p.composeFilePath)
		if err != nil {
			return nil, fmt.Errorf("error parsing docker-compose file: %w", err)
		}
	}

	return p.composeFile, nil
}

// AddDefaultBuildService adds a service to the list of default build services
func (p *DockerComposeInfo) AddDefaultBuildService(service string) {
	p.defaultPushImages = append(p.defaultBuildServices, service)
}

// AddDefaultBuildServices adds multiple services to the list of default build services
func (p *DockerComposeInfo) AddDefaultBuildServices(services []string) {
	p.defaultPushImages = append(p.defaultBuildServices, services...)
}

// getDefaultBuildServices returns the default build services (if the user does not provide any default build service, returns all buildable
// services)
func (p *DockerComposeInfo) GetDefaultBuildServices() ([]string, error) {
	if len(p.defaultBuildServices) == 0 {
		file, err := p.GetComposeFile()
		if err != nil {
			return nil, fmt.Errorf("error getting compose file: %w", err)
		}

		for serviceName, service := range file.Services {
			if service.Build != nil { // Must be a buildable service
				p.defaultBuildServices = append(p.defaultBuildServices, serviceName)
			}
		}
	}

	return p.defaultBuildServices, nil
}

// AddDefaultPushImage adds an image to the list of default push images
func (p *DockerComposeInfo) AddDefaultPushImage(image string) {
	p.defaultPushImages = append(p.defaultPushImages, image)
}

// AddDefaultPushImages adds multiple images to the list of default push images
func (p *DockerComposeInfo) AddDefaultPushImages(images []string) {
	p.defaultPushImages = append(p.defaultPushImages, images...)
}

// GetDefaultPushImages returns the default images to be pushed (if the user does not provide any default push image, uses the images of all
// buildable services that have an image name defined)
func (p *DockerComposeInfo) GetDefaultPushImages() ([]string, error) {
	if len(p.defaultPushImages) == 0 {
		file, err := p.GetComposeFile()
		if err != nil {
			return nil, fmt.Errorf("error getting compose file: %w", err)
		}

		for _, service := range file.Services {
			if service.Build != nil && service.Image != nil { // Must be a buildable service and have an image name defined
				p.defaultPushImages = append(p.defaultPushImages, *service.Image)
			}
		}
	}

	return p.defaultPushImages, nil
}

// GetPushImages returns the list of default push images. You must not modify this list
func (p *DockerComposeInfo) GetPushImages() []string {
	return p.defaultPushImages
}

// GetServiceImage returns the image of a service
func (p *DockerComposeInfo) GetServiceImage(serviceName string) (string, error) {
	file, err := p.GetComposeFile()
	if err != nil {
		return "", fmt.Errorf("error getting compose file: %w", err)
	}

	service, ok := file.Services[serviceName]
	if !ok {
		return "", fmt.Errorf("service %s not found", serviceName)
	}

	return *service.Image, nil
}
