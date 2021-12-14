package notifications

import (
	"github.com/gen2brain/beeep"
)

type Notification struct {
	Title, Message string
}

type notificationsHandler struct {
	showNotifications bool
	notifications     chan Notification
}

func (nh *notificationsHandler) show() {
	for n := range nh.notifications {
		if nh.showNotifications {
			go beeep.Notify(n.Title, n.Message, "")
		}
	}
}

// On turns notifications on
func (nh *notificationsHandler) On() {
	nh.showNotifications = true
}

// Off turns notifications off
func (nh *notificationsHandler) Off() {
	nh.showNotifications = false
}

// Send will generate an appropriately formatted notification and send it
// through the notification channel if notifications are enabled
func (nh *notificationsHandler) Send(title string, message string) {
	nh.notifications <- Notification{
		Title:   title,
		Message: message,
	}
}

// NewNotificationsHandler creates a new NotificationsHandler to
// toggle the showing and format and display notifications for the app
func NewNotificationsHandler() *notificationsHandler {
	n := &notificationsHandler{
		showNotifications: false,
		notifications:     make(chan Notification),
	}
	go n.show()
	return n
}
