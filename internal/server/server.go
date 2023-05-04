// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package server

import (
	"github.com/joshuar/autocorrector/internal/control"
	"github.com/joshuar/autocorrector/internal/keytracker"
	log "github.com/sirupsen/logrus"
)

func Run(user string) {
	keyTracker := keytracker.NewKeyTracker()

	for {
		socket := control.CreateServer(user)
		go func() {
			for w := range keyTracker.WordToCheck {
				socket.SendWord(&control.WordMsg{Word: w})
			}
		}()
		for {
			select {
			case msg := <-socket.Data:
				switch t := msg.(type) {
				case *control.StateMsg:
					switch t.State {
					case control.Pause:
						keyTracker.Ctrl <- true
					case control.Resume:
						keyTracker.Ctrl <- false
					default:
						log.Debugf("Unknown state: %v", msg)
					}
				case *control.WordMsg:
					keyTracker.CorrectionToMake <- t
				default:
					log.Debugf("Unknown message %T received: %v", msg, msg)
				}
			case <-socket.Done:
				log.Debug("Received done, restarting socket...")
				socket = control.CreateServer(user)
			}
		}
	}

}
