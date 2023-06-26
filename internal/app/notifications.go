package app

import (
	"fyne.io/fyne/v2"
)

// NewNotificationsHandler creates a new NotificationsHandler to
// toggle the showing and format and display notifications for the app
func (a *App) notificationHandler() chan fyne.Notification {
	notificationsCh := make(chan fyne.Notification)
	go func() {
		for notification := range notificationsCh {
			if a.showNotifications {
				a.app.SendNotification(&notification)
			}
		}
	}()
	return notificationsCh
}
