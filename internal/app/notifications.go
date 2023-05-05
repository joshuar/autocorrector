package app

import (
	"fyne.io/fyne/v2"
	"github.com/joshuar/autocorrector/internal/notifications"
	"github.com/rs/zerolog/log"
)

type notificationsHandler struct {
	showNotifications bool
	data              chan interface{}
}

// NewNotificationsHandler creates a new NotificationsHandler to
// toggle the showing and format and display notifications for the app
func (a *App) setupNotifications() {
	a.notifyHandler = &notificationsHandler{
		showNotifications: false,
		data:              make(chan interface{}),
	}
	go func() {
		for n := range a.notifyHandler.data {
			switch n := n.(type) {
			case bool:
				a.notifyHandler.showNotifications = n
			case notifications.Notification:
				if a.notifyHandler.showNotifications {
					a.app.SendNotification(&fyne.Notification{
						Title:   n.Title,
						Content: n.Message,
					})
				}
			default:
				log.Debug().Caller().
					Msgf("Unexpected data %T on notification channel: %v", n, n)
			}
		}
	}()
}
