//go:build mage

package main

import (
	"errors"
	"strings"

	"github.com/magefile/mage/sh"
)

func Build() error {
	return sh.RunV("go", "build", "-o", "bin/api", "./cmd/api")
}

func Test() error {
	return sh.RunV("go", "test", "./...")
}

func Clean() error {
	return sh.RunV("rm", "-rf", "bin")
}

func Deploy() error {
	err := sh.RunV("sudo", "go", "run", "./scripts/dagger.go")
	if err != nil {
		return err
	}
	output, err := sh.Output("sudo", "docker", "load", "-i", "/tmp/api.tar")
	if err != nil {
		return err
	}
	imageID, err := extractImageID(output)
	if err != nil {
		return err
	}
	err = sh.RunV("sudo", "docker", "tag", imageID, "api:latest")
	if err != nil {
		return err
	}
	err = sh.RunV("sudo", "docker", "compose", "up", "-d")
	if err != nil {
		return err
	}
	return nil
}

func extractImageID(output string) (string, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Loaded image ID: ") {
			parts := strings.Split(line, ":")
			if len(parts) == 3 {
				return parts[2], nil
			}
		}
	}
	return "", errors.New("image ID not found in docker load output")
}
