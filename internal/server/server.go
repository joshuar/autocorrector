// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package server

import (
	"context"

	"github.com/joshuar/autocorrector/internal/keytracker"
)

func Run(user string) {
	ctx := context.Background()
	keytracker.NewKeyTracker(ctx)

	<-ctx.Done()

	// for {
	// 	socket := control.CreateServer(user)
	// 	go func() {
	// 		for w := range keyTracker.WordCh {
	// 			if !keyTracker.Paused() {
	// 				socket.SendWord(&control.WordMsg{Word: w})
	// 			}
	// 		}
	// 	}()
	// 	for {
	// 		select {
	// 		case msg := <-socket.Data:
	// 			switch t := msg.(type) {
	// 			case *control.StateMsg:
	// 				switch t.State {
	// 				case control.Pause:
	// 					keyTracker.ControlCh <- true
	// 				case control.Resume:
	// 					keyTracker.ControlCh <- false
	// 				default:
	// 					log.Debug().Msgf("Unknown state: %v", msg)
	// 				}
	// 			case *control.WordMsg:
	// 				keyTracker.CorrectionCh <- t
	// 			default:
	// 				log.Debug().Msgf("Unknown message %T received: %v", msg, msg)
	// 			}
	// 		case <-socket.Done:
	// 			log.Debug().Msg("Received done, restarting socket...")
	// 			keyTracker.ControlCh <- true
	// 			socket = control.CreateServer(user)
	// 		}
	// 	}
	// }

}
