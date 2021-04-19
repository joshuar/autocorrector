package control

import (
	"encoding/gob"
	"io"
	"net"
	"os"
	"os/user"
	"strconv"

	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
)

const (
	Acknowledge        MsgType = 00
	PauseServer        MsgType = 11
	ResumeServer       MsgType = 12
	StartServer        MsgType = 18
	StopServer         MsgType = 19
	CorrectionFound    MsgType = 21
	CorrectionNotFound MsgType = 22
	HideNotifications  MsgType = 31
	ShowNotifications  MsgType = 32
	Notification       MsgType = 33
	GetStats           MsgType = 41
	ServerStarted      MsgType = 98
	ServerStopped      MsgType = 99
)

type MsgType int
type ControlMessage struct {
	Type MsgType
	Data interface{}
}

type NotificationData struct {
	Title, Message string
}

type ControlSocket struct {
	recevierName net.Listener
	conn         io.ReadWriteCloser
	receivePath  string
	sendPath     string
	Data         chan *ControlMessage
}

func (s *ControlSocket) AcceptConnections(manager *ConnManager) {
	for {
		conn, err := s.recevierName.Accept()
		if err != nil {
			log.Fatalf("Error on accept: %s", err)
		}
		s.conn = conn
		manager.register <- s
		gob.Register(NotificationData{})
		go s.recieveMessage()
	}
}

func (s *ControlSocket) recieveMessage() {
	dec := gob.NewDecoder(s.conn)
	var m *ControlMessage
	if err := dec.Decode(&m); err != nil {
		log.Errorf("Error on read: %s", err)
	}
	s.Data <- m
}

func (s *ControlSocket) SendMessage(msgType MsgType, msgData interface{}) {
	tryMessage := func() error {
		c, err := net.Dial("unix", s.sendPath)
		if err != nil {
			log.Errorf("Failed to dial: %s", err)
			return err
		} else {
			defer c.Close()
			enc := gob.NewEncoder(c)
			message := &ControlMessage{
				Type: msgType,
				Data: msgData,
			}
			if err := enc.Encode(message); err != nil {
				log.Errorf("Write error: %s", err)
			}
			log.Debugf("Sent message: %v: %v", message.Type, message.Data)
			return nil
		}
	}
	err := backoff.Retry(tryMessage, backoff.NewExponentialBackOff())
	if err != nil {
		log.Errorf("Problem with connection backoff", err)
	}
}

func NewSocket(username string) *ControlSocket {
	var u *user.User
	var err error
	var receivePath, sendPath string
	if username != "" {
		u, err = user.Lookup(username)
		receivePath = "/tmp/autocorrector-" + u.Username + "-server.sock"
		sendPath = "/tmp/autocorrector-" + u.Username + "-client.sock"
	} else {
		u, err = user.Current()
		receivePath = "/tmp/autocorrector-" + u.Username + "-client.sock"
		sendPath = "/tmp/autocorrector-" + u.Username + "-server.sock"
	}
	if err != nil {
		log.Fatal(err)
	}
	s := &ControlSocket{
		receivePath: receivePath,
		sendPath:    sendPath,
		Data:        make(chan *ControlMessage),
	}
	s.recevierName = listen(s.receivePath)
	if username != "" {
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)
		if err := os.Chown(s.receivePath, uid, gid); err != nil {
			log.Fatalf("Unable to change ownership on socket file %s to %s:%s : %s", s.receivePath, 0, gid, err)
		}
	}
	return s
}

func listen(socketPath string) net.Listener {
	os.Remove(socketPath)
	name, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Unable to listen on socket file %s: %s", socketPath, err)
	}
	return name
}

type ConnManager struct {
	clients map[*ControlSocket]bool
	// broadcast  chan *ControlMessage
	register   chan *ControlSocket
	unregister chan *ControlSocket
}

func (manager *ConnManager) Start() {
	for {
		select {
		case connection := <-manager.register:
			manager.clients[connection] = true
			log.Debug("New connection")
		case connection := <-manager.unregister:
			if _, ok := manager.clients[connection]; ok {
				close(connection.Data)
				delete(manager.clients, connection)
				log.Debug("Connection ended")
			}
			// case message := <-manager.broadcast:
			// 	for connection := range manager.clients {
			// 		select {
			// 		case connection.Data <- message:
			// 		default:
			// 			close(connection.Data)
			// 			delete(manager.clients, connection)
			// 		}
			// 	}
		}
	}
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		clients: make(map[*ControlSocket]bool),
		// broadcast:  make(chan []byte),
		register:   make(chan *ControlSocket),
		unregister: make(chan *ControlSocket),
	}
}
