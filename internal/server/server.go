// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package server

import (
	"context"

	"fyne.io/fyne/v2"
	"github.com/joshuar/autocorrector/internal/keytracker"
)

func Run(ctx context.Context, notificationsCh chan fyne.Notification) {
	keytracker.NewKeyTracker(ctx, notificationsCh)
	<-ctx.Done()
}
