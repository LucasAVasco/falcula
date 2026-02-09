// Package cmd implements docker-compose commands. The function have the same name as the docker-compose command
package cmd

import (
	"path/filepath"

	"github.com/LucasAVasco/falcula/process"
)

func genericDockerComposeCommand(opts *process.Options, composeFile, command string, args ...string) (*process.Process, error) {
	dir := filepath.Dir(composeFile)

	newArgs := make([]string, 0, len(args)+7)

	newArgs = append(newArgs, "compose", "--project-directory", dir, "-f", composeFile, command)
	newArgs = append(newArgs, args...)

	return process.New(opts, "docker", newArgs...)
}

func genericDockerCommand(opts *process.Options, args ...string) (*process.Process, error) {
	return process.New(opts, "docker", args...)
}

// Pull pulls images referenced in the compose file. Does not pull buildable images
func Pull(opts *process.Options, composeFile string) (*process.Process, error) {
	return genericDockerComposeCommand(opts, composeFile, "pull", "--ignore-buildable")
}

func Build(opts *process.Options, composeFile string, args ...string) (*process.Process, error) {
	return genericDockerComposeCommand(opts, composeFile, "build", args...)
}

func Up(opts *process.Options, composeFile string) (*process.Process, error) {
	return genericDockerComposeCommand(opts, composeFile, "up")
}

func Stop(opts *process.Options, composeFile string) (*process.Process, error) {
	return genericDockerComposeCommand(opts, composeFile, "stop")
}

func Down(opts *process.Options, composeFile string) (*process.Process, error) {
	return genericDockerComposeCommand(opts, composeFile, "down")
}

func Kill(opts *process.Options, composeFile string) (*process.Process, error) {
	return genericDockerComposeCommand(opts, composeFile, "kill")
}

func Tag(opts *process.Options, image string, tag string) (*process.Process, error) {
	return genericDockerCommand(opts, "tag", image, tag)
}

// Push pushes images referenced in the compose file to a registry. The registry URL is automatically prefixed to the image, unless it is
// empty
func Push(opts *process.Options, image string, registry string) (*process.Process, error) {
	if registry != "" {
		image = registry + "/" + image
	}
	return genericDockerCommand(opts, "push", image)
}

// ManifestCreate creates a manifest from a list of images. The registry URL is automatically prefixed to the image, unless it is empty
func ManifestCreate(opts *process.Options, manifest string, registry string, images ...string) (*process.Process, error) {
	if registry != "" {
		manifest = registry + "/" + manifest
	}
	cmdArgs := []string{"manifest", "create", manifest}

	for _, image := range images {
		if registry != "" {
			image = registry + "/" + image
		}
		cmdArgs = append(cmdArgs, "--amend", image)
	}

	return genericDockerCommand(opts, cmdArgs...)
}
