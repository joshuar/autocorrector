// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"syscall"

	"github.com/joshuar/autocorrector/cmd"
	"github.com/rs/zerolog/log"
)

func main() {
	cmd.Execute()
}

// Following is copied from https://git.kernel.org/pub/scm/libs/libcap/libcap.git/tree/goapps/web/web.go
//
// ensureNotEUID aborts the program if it is running setuid something,
// or being invoked by root.
func init() {
	euid := syscall.Geteuid()
	uid := syscall.Getuid()
	egid := syscall.Getegid()
	gid := syscall.Getgid()
	if uid != euid || gid != egid || uid == 0 {
		log.Fatal().Msg("autocorrector should not be run with additional privileges or as root.")
	}
}
