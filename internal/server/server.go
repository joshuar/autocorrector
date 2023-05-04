// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package server

import (
	"fmt"

	"github.com/joshuar/autocorrector/internal/control"
	"github.com/joshuar/autocorrector/internal/corrections"
	"github.com/joshuar/autocorrector/internal/notifications"
	"github.com/joshuar/autocorrector/internal/wordstats"
	log "github.com/sirupsen/logrus"
)

func Start() {
	socket := control.CreateClient()
	notifyCtrl := notifications.NewNotificationsHandler()
	stats := wordstats.RunStats()
	corrections := corrections.NewCorrections()

	handleSocket := func() {
		for msg := range socket.Data {
			// case: recieved data on the socket
			switch t := msg.(type) {
			case *control.WordMsg:
				stats.Checked <- t.Word
				correction, found := corrections.CheckWord(t.Word)
				if found {
					t.Correction = correction
					stats.Corrected <- [2]string{t.Word, t.Correction}
					notifyCtrl <- notifications.Notification{
						Title:   "Correction!",
						Message: fmt.Sprintf("Corrected %s with %s", t.Word, t.Correction),
					}
				}
				socket.SendWord(t)
			default:
				log.Debugf("Unknown message %T received: %v", msg, msg)
			}
		}
	}

	go handleSocket()
}
