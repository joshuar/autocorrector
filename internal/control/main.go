package control

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/user"
	"strconv"

	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
)

type StateMsg struct {
	Start, Stop, Pause, Resume bool
}

type WordMsg struct {
	Word, Correction string
}

type Msg struct {
	*StateMsg
	*WordMsg
}

type ControlSocket struct {
	localSocket net.Listener
	localPath   string
	remotePath  string
}

func (s *ControlSocket) AcceptConnections(r chan net.Conn) {
	for {
		conn, err := s.localSocket.Accept()
		if err != nil {
			log.Fatalf("Error on accept: %s", err)
		}
		r <- conn
	}
}

func NewSocketConnection(username string) *ControlSocket {
	var u *user.User
	var err error
	var local, remote string
	if username != "" {
		u, err = user.Lookup(username)
		local = "/tmp/autocorrector-" + u.Username + "-server.sock"
		remote = "/tmp/autocorrector-" + u.Username + "-client.sock"
	} else {
		u, err = user.Current()
		local = "/tmp/autocorrector-" + u.Username + "-client.sock"
		remote = "/tmp/autocorrector-" + u.Username + "-server.sock"
	}
	if err != nil {
		log.Fatal(err)
	}
	if err := os.Remove(local); err != nil {
		log.Fatalf("Could not remove existing socket path: %s (%v)", local, err)

	}
	l, err := net.Listen("unix", local)
	if err != nil {
		log.Fatalf("Unable to listen on socket file %s: %s", local, err)
	}
	s := &ControlSocket{
		localSocket: l,
		localPath:   local,
		remotePath:  remote,
	}
	if username != "" {
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)
		if err := os.Chown(s.localPath, uid, gid); err != nil {
			log.Fatalf("Unable to change ownership on socket file %s to %s:%s : %s", s.localPath, 0, gid, err)
		}
	}
	return s
}

type ConnManager struct {
	socket *ControlSocket
	connCh chan net.Conn
	Data   chan interface{}
}

func (m *ConnManager) Start() {
	go m.socket.AcceptConnections(m.connCh)
	for c := range m.connCh {
		go m.RecieveMessage(c)
	}
}

func (m *ConnManager) RecieveMessage(connection net.Conn) {
	var rawMsg Msg
	dec := gob.NewDecoder(connection)
	if err := dec.Decode(&rawMsg); err == nil {
		switch {
		case rawMsg.StateMsg != nil:
			m.Data <- rawMsg.StateMsg
		case rawMsg.WordMsg != nil:
			m.Data <- rawMsg.WordMsg
		default:
			log.Errorf("Decoded but unhandled data received: %v", rawMsg)
		}
	}
	connection.Close()
}

func (manager *ConnManager) SendState(state *StateMsg) {
	manager.SendMessage(state)
}

func (manager *ConnManager) SendWord(w string, c string) {
	t := &WordMsg{
		Word:       w,
		Correction: c,
	}
	manager.SendMessage(t)
}

func (manager *ConnManager) SendMessage(msgData interface{}) {
	tryMessage := func() error {
		conn, err := net.Dial("unix", manager.socket.remotePath)
		if err != nil {
			log.Errorf("Failed to dial: %s", err)
			return err
		}
		enc := gob.NewEncoder(conn)
		switch t := msgData.(type) {
		case *StateMsg:
			if err := enc.Encode(&Msg{StateMsg: t}); err != nil {
				log.Errorf("Write error: %s", err)
				return err
			}
		case *WordMsg:
			if err := enc.Encode(&Msg{WordMsg: t}); err != nil {
				log.Errorf("Write error: %s", err)
				return err
			}
		default:
			return fmt.Errorf("unknown data to send: %v", t)
		}
		return nil
	}
	err := backoff.Retry(tryMessage, backoff.NewExponentialBackOff())
	if err != nil {
		log.Errorf("Problem with connection backoff", err)
	}
}

func NewConnManager(user string) *ConnManager {
	var s *ControlSocket
	if user != "" {
		s = NewSocketConnection(user)
	} else {
		s = NewSocketConnection("")
	}
	return &ConnManager{
		socket: s,
		Data:   make(chan interface{}),
		connCh: make(chan net.Conn),
	}
}