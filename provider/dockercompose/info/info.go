// Package info provides information about a docker-compose file
package info

import ()

// DockerComposeInfo holds information about a docker-compose file
type DockerComposeInfo struct {
	composeFilePath      string
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

// AddDefaultBuildService adds a service to the list of default build services
func (p *DockerComposeInfo) AddDefaultBuildService(service string) {
	p.defaultPushImages = append(p.defaultBuildServices, service)
}

// AddDefaultBuildServices adds multiple services to the list of default build services
func (p *DockerComposeInfo) AddDefaultBuildServices(services []string) {
	p.defaultPushImages = append(p.defaultBuildServices, services...)
}

// getDefaultBuildServices returns the default build services (if the user does not provide any build service, these will be used)
func (p *DockerComposeInfo) GetDefaultBuildServices() []string {
	return p.defaultBuildServices
}

// AddDefaultPushImage adds an image to the list of default push images
func (p *DockerComposeInfo) AddDefaultPushImage(image string) {
	p.defaultPushImages = append(p.defaultPushImages, image)
}

// AddDefaultPushImages adds multiple images to the list of default push images
func (p *DockerComposeInfo) AddDefaultPushImages(images []string) {
	p.defaultPushImages = append(p.defaultPushImages, images...)
}

// GetDefaultPushImages returns the default images to be pushed (if the user does not provide any image, these will be used)
func (p *DockerComposeInfo) GetDefaultPushImages() []string {
	return p.defaultPushImages
}

// GetPushImages returns the list of default push images. You must not modify this list
func (p *DockerComposeInfo) GetPushImages() []string {
	return p.defaultPushImages
}
