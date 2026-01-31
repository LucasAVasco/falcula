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

func Build(opts *process.Options, composeFile string) (*process.Process, error) {
	return genericDockerComposeCommand(opts, composeFile, "build")
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

func Push(opts *process.Options, image string, repository string) (*process.Process, error) {
	return genericDockerCommand(opts, "push", repository+"/"+image)
}
