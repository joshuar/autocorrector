// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"net/http"
	"os"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func setLogging() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func setDebugging() {
	if debugFlag {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled.")
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func setProfiling() {
	if profileFlag {
		go func() {
			log.Debug().Err(http.ListenAndServe("localhost:6061", nil))
		}()
		log.Debug().Msg("Profiling is enabled and available at localhost:6061")
	}
}

func ensureEUID() {
	euid := syscall.Geteuid()
	uid := syscall.Getuid()
	egid := syscall.Getegid()
	// gid := syscall.Getgid()
	if euid != 0 || egid != 0 || uid != 0 {
		log.Fatal().Msg("autocorrector server must be run as root.")
	}
}
