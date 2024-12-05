//go:build mage

package main

import (
	"github.com/carolynvs/magex/pkg"
	"github.com/magefile/mage/sh"
)

var Default = Build

func Build() error {
	return sh.RunV("go", "build", "-o", "bin/api", "./cmd/api")
}

func Test() error {
	return sh.RunV("go", "test", "./...")
}

func Clean() error {
	return sh.RunV("rm", "-rf", "bin")
}

func Tidy() error {
	return sh.RunV("go", "mod", "tidy")
}

func DeployImage() error {
	return sh.RunV("sudo", "docker", "compose", "-f", "deployments/docker-compose.yaml", "up", "-d")
}

func BuildImage() error {
	return sh.RunV("sudo", "docker", "build", "-t", "api:latest", ".")
}

func BuildAndDeploy() error {
	err := BuildImage()
	if err != nil {
		return err
	}
	return DeployImage()
}

func EnsureMage() error {
	return pkg.EnsureMage("")
}
