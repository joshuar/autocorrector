//+build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Runs go mod download and then installs the binary.
func Build() error {
	mg.Deps(installDevelopmentDependancies)
	if err := sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	if err := sh.Run("go", "build"); err != nil {
		return err
	}
	return nil
}

func installDevelopmentDependancies() {
	sh.Run("sudo", "dnf", "-y", "install", "gtk3-devel", "libappindicator-gtk3-devel", "libevdev-devel")
}
