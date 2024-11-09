//go:build mage

package main

import (
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
	var err error
	err = sh.RunV("sudo", "docker", "build", "-t", "api:latest", ".")
	if err != nil {
		return err
	}
	err = sh.RunV("sudo", "docker", "compose", "-f", "deployments/docker-compose.yaml", "up", "-d")
	if err != nil {
		return err
	}
	return nil
}
