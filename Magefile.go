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
