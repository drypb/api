//go:build mage

package main

import "github.com/magefile/mage/sh"

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
	err := sh.Run("sudo", "docker", "build", "-t", "api:latest", ".")
	if err != nil {
		return err
	}
	err = sh.Run("sudo", "docker", "compose", "up", "-d")
	if err != nil {
		return err
	}
	return nil
}
