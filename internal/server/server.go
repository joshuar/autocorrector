// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package server

import (
	"context"

	"github.com/joshuar/autocorrector/internal/keytracker"
)

func Run(ctx context.Context) {
	keytracker.NewKeyTracker(ctx)
	<-ctx.Done()
}
