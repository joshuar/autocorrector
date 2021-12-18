package notifications

import (
	"github.com/gen2brain/beeep"
	log "github.com/sirupsen/logrus"
)

type Notification struct {
	Title, Message string
}

type notificationsHandler struct {
	showNotifications bool
	data              chan interface{}
}

func (nh *notificationsHandler) handler() {
	for n := range nh.data {
		switch n := n.(type) {
		case bool:
			nh.showNotifications = n
		case Notification:
			if nh.showNotifications {
				go beeep.Notify(n.Title, n.Message, "")
			}
		default:
			log.Debug("Unexpected data %T on notification channel: %v", n, n)
		}
	}
}

// NewNotificationsHandler creates a new NotificationsHandler to
// toggle the showing and format and display notifications for the app
func NewNotificationsHandler() *chan interface{} {
	n := &notificationsHandler{
		showNotifications: false,
		data:              make(chan interface{}),
	}
	go n.handler()
	return &n.data
}
