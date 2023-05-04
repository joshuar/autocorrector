// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"net/http"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func setDebugging() {
	if debugFlag {
		log.SetLevel(log.DebugLevel)
	}
}

func setProfiling() {
	if profileFlag {
		go func() {
			log.Info(http.ListenAndServe("localhost:6061", nil))
		}()
		log.Info("Profiling is enabled and available at localhost:6061")
	}
}

func ensureEUID() {
	euid := syscall.Geteuid()
	uid := syscall.Getuid()
	egid := syscall.Getegid()
	// gid := syscall.Getgid()
	if euid != 0 || egid != 0 || uid != 0 {
		log.Fatalf("autocorrector server must be run as root.")
	}
}
