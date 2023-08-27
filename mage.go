// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:build mage

package main

import (
	"github.com/magefile/mage/sh"
)

var Default = Build

func Build() error {
	if err := sh.Run("go", "build"); err != nil {
		return err
	}
	if err := sh.Run("sudo", "setcap", "cap_setgid,cap_setuid=p", "autocorrector"); err != nil {
		return err
	}
	return nil
}

func Clean() error {
	return sh.Run("rm", "autocorrector")
}
