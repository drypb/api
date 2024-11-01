//go:build mage

package main

import (
	"errors"
	"strings"

	"github.com/magefile/mage/sh"
)

func Build() error {
	err := sh.Run("go", "build", "-o", "bin/api", "./cmd/api")
	if err != nil {
		return err
	}
	return nil
}

func Test() error {
	err := sh.Run("go", "test", "./...")
	if err != nil {
		return err
	}
	return nil
}

func Clean() error {
	err := sh.Run("rm", "-rf", "bin")
	if err != nil {
		return err
	}
	return nil
}

func Deploy() error {
	err := sh.RunV("go", "run", "dagger.go")
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
