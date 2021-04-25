//+build mage

package main

import (
	"os"
	"os/user"

	"text/template"

	"github.com/adrg/xdg"
	"github.com/magefile/mage/sh"
)

// Runs go mod download and then installs the binary.
func Build() error {
	tmplVars := make(map[string]interface{})

	if err := sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	if err := sh.Run("go", "build"); err != nil {
		return err
	}
	u, err := user.Current()
	tmplVars["User"] = u.Username
	if err != nil {
		return err
	}
	t, err := template.ParseFiles("autocorrector-server.service.tmpl")
	if err != nil {
		return err
	}
	f, err := os.Create("autocorrector-server.service")
	if err := t.Execute(f, tmplVars); err != nil {
		return err
	}
	return nil
}

// Installs the binary and the systemd files
func Install() error {

	systemdServerServiceFile := "/etc/systemd/system/autocorrector-server.service"
	systemdClientServiceFile, err := xdg.ConfigFile("systemd/user/autocorrector-client.service")
	if err != nil {
		return err
	}
	if err := sh.Copy("autocorrector-server.service", systemdServerServiceFile); err != nil {
		return err
	}
	if err := sh.Copy("autocorrector-client.service", systemdClientServiceFile); err != nil {
		return err
	}
	return nil
}
