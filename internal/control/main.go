package control

import (
	"encoding/gob"
	"io"
	"net"
	"os"
	"os/user"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
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

type ControlSocket struct {
	recevierName net.Listener
	conn         io.ReadWriteCloser
	receivePath  string
	sendPath     string
	User         *user.User
	Data         chan ControlMessage
}

type MsgType int
type ControlMessage struct {
	Type MsgType
	Data interface{}
}

type NotificationData struct {
	Title, Message string
}

func (s *ControlSocket) AcceptConnections() {
	for {
		conn, err := s.recevierName.Accept()
		if err != nil {
			log.Fatalf("Error on accept: %s", err)
		}
		s.conn = conn
		gob.Register(NotificationData{})
		go s.recieveMessage()
	}
}

func NewServerSocket(username string) *ControlSocket {
	var u *user.User
	var err error
	u, err = user.Lookup(username)
	if err != nil {
		log.Fatal(err)
	}
	s := &ControlSocket{
		receivePath: "/tmp/autocorrector-" + u.Username + "-server.sock",
		sendPath:    "/tmp/autocorrector-" + u.Username + "-client.sock",
		User:        u,
		Data:        make(chan ControlMessage),
	}
	s.recevierName = listen(s.receivePath)
	uid, _ := strconv.Atoi(s.User.Uid)
	gid, _ := strconv.Atoi(s.User.Gid)
	if err := os.Chown(s.receivePath, uid, gid); err != nil {
		log.Fatalf("Unable to change ownership on socket file %s to %s:%s : %s", s.receivePath, 0, gid, err)
	}
	return s
}

func NewClientSocket() *ControlSocket {
	var u *user.User
	var err error
	u, err = user.Current()
	if err != nil {
		log.Fatal(err)
	}
	s := &ControlSocket{
		receivePath: "/tmp/autocorrector-" + u.Username + "-client.sock",
		sendPath:    "/tmp/autocorrector-" + u.Username + "-server.sock",
		User:        u,
		Data:        make(chan ControlMessage),
	}
	s.recevierName = listen(s.receivePath)
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

func (s *ControlSocket) recieveMessage() {
	dec := gob.NewDecoder(s.conn)
	var m ControlMessage
	if err := dec.Decode(&m); err != nil {
		log.Errorf("Error on read: %s", err)
	}
	s.Data <- m
}

func (s *ControlSocket) SendMessage(msgType MsgType, msgData interface{}) {
	c, err := net.Dial("unix", s.sendPath)
	if err != nil {
		log.Errorf("Failed to dial: %s", err)

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
		log.Debugf("Sent message: %v", message)
	}
}
