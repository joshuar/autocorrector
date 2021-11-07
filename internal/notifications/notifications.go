package notifications

import (
	"github.com/gen2brain/beeep"
)

type notificationsHandler struct {
	ShowNotifications bool
}

// Send will generate an appropriately formatted notification and send it through the notification channel if notifications are enabled
func (nh *notificationsHandler) Send(title string, message string) {
	if nh.ShowNotifications {
		beeep.Notify(title, message, "")
	}
}

// NewNotificationsHandler will create a struct that opens a channel used for sending/receiving notifications and a bool to track whether they should be displayed
func NewNotificationsHandler() *notificationsHandler {
	return &notificationsHandler{
		ShowNotifications: false,
	}
}
