/*
docker.go contains all of the logic specific to executing docker processes
from inside the application
*/
package main

import (
	"os/exec"
	"path/filepath"
)

/*
execDocker is a pretty wrapper around exec.Command("docker",...)
*/
func execDocker(path string, command string, args ...string) (output string, err error) {
	// Hold the output from our command
	var outputBytes []byte

	// First, we create an array with command and args to pass to exec
	tmp := []string{command}
	for _, arg := range args {
		tmp = append(tmp, arg)
	}

	// Next, we build and execute the command
	cmd := exec.Command("docker", tmp...)

	cmd.Dir, err = filepath.Abs(path)
	if err != nil {
		return
	}

	outputBytes, err = cmd.CombinedOutput()
	output = string(outputBytes)

	return
}

type DockerOpts struct {
	Cache bool
}

/*
buildImage will take a path to a docker image, and execute docker build as a
child process. It will tag the docker built image as name, this allows us to
later build other images using this one as a base. It captures stdout and
stderr returning them both in output.
*/
func buildImage(name string, path string, opts DockerOpts) (output string, err error) {
	args := []string{"-t", name}

	if !opts.Cache {
		args = append(args, "--no-cache")
	}

	// local directory
	args = append(args, ".")

	return execDocker(path, "build", args...)
}

/*
pushImage will take a docker image and push it to a remote registry. It captures
stdout and stderr returning them both in output
*/
func pushImage(name string) (output string, err error) {
	return execDocker("/", "push", name)
}

func dockerAlias(name string, alias string) (output string, err error) {
	return execDocker("/", "tag", "-f", name, alias)
}
