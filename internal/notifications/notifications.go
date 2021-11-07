package notifications

import (
	"github.com/gen2brain/beeep"
	log "github.com/sirupsen/logrus"
)

type notificationMsg struct {
	title, message string
}
type notificationsHandler struct {
	ShowCorrections bool
	notification    chan *notificationMsg
}

// Show will generate an appropriately formatted notification and send it through the notification channel if notifications are enabled
func (nh *notificationsHandler) Show(t string, m string) {
	n := &notificationMsg{
		title:   t,
		message: m,
	}
	nh.notification <- n
}

// NewNotificationsHandler will create a struct that opens a channel used for sending/receiving notifications and a bool to track whether they should be displayed
func NewNotificationsHandler() *notificationsHandler {
	nh := &notificationsHandler{
		ShowCorrections: false,
		notification:    make(chan *notificationMsg),
	}
	go func() {
		for n := range nh.notification {
			log.Debugf("Recieved notification %s: %s", n.title, n.message)
			if nh.ShowCorrections {
				beeep.Notify(n.title, n.message, "")
			}
		}
	}()
	return nh
}
