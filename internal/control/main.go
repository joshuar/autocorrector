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

type StatsMsg struct {
	Word, Correction string
}

type Msg struct {
	*StateMsg
	*StatsMsg
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
	s := &ControlSocket{
		localSocket: listen(local),
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

func listen(socketPath string) net.Listener {
	os.Remove(socketPath)
	name, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Unable to listen on socket file %s: %s", socketPath, err)
	}
	return name
}

type ConnManager struct {
	socket     *ControlSocket
	register   chan net.Conn
	unregister chan net.Conn
	Data       chan interface{}
}

func (manager *ConnManager) Start() {
	go manager.socket.AcceptConnections(manager.register)
	for {
		select {
		case connection := <-manager.register:
			go manager.RecieveMessage(connection)
		case connection := <-manager.unregister:
			connection.Close()
		}
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
		socket:     s,
		Data:       make(chan interface{}),
		register:   make(chan net.Conn),
		unregister: make(chan net.Conn),
	}
}

func (manager *ConnManager) RecieveMessage(connection net.Conn) {
	var m Msg
	dec := gob.NewDecoder(connection)
	if err := dec.Decode(&m); err == nil {
		log.Debugf("Decoded msg: %v", m)
		switch {
		case m.StateMsg != nil:
			manager.Data <- m.StateMsg
		case m.StatsMsg != nil:
			manager.Data <- m.StatsMsg
		}
	}
	manager.unregister <- connection
}

func (manager *ConnManager) SendState(state *StateMsg) {
	manager.SendMessage(state)
}

func (manager *ConnManager) SendStats(stats *StatsMsg) {
	manager.SendMessage(stats)
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
		case *StatsMsg:
			if err := enc.Encode(&Msg{StatsMsg: t}); err != nil {
				log.Errorf("Write error: %s", err)
				return err
			}
		default:
			log.Debugf("Got %v", t)
			return fmt.Errorf("unknown data to send: %v", t)
		}
		return nil
	}
	err := backoff.Retry(tryMessage, backoff.NewExponentialBackOff())
	if err != nil {
		log.Errorf("Problem with connection backoff", err)
	}
}
