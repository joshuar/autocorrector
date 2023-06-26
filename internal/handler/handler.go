// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handler

import (
	"fyne.io/fyne/v2"
	"github.com/joshuar/autocorrector/internal/word"
)

type handler struct {
	WordCh          chan word.WordDetails
	CorrectionCh    chan word.WordDetails
	NotificationsCh chan fyne.Notification
}

func NewHandler() *handler {
	return &handler{
		WordCh:          make(chan word.WordDetails),
		CorrectionCh:    make(chan word.WordDetails),
		NotificationsCh: make(chan fyne.Notification),
	}
}
