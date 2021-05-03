package control

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"strconv"

	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
)

const (
	SocketPath       = "/tmp/autocorrector.sock"
	Start      State = 0x01
	Resume     State = 0x02
	Pause      State = 0x03
	Stop       State = 0x04
)

type State int8

type StateMsg struct {
	State
}

type WordMsg struct {
	Word, Correction string
	Punct            rune
}

type Msg struct {
	*StateMsg
	*WordMsg
}

type Socket struct {
	addr     *net.UnixAddr
	listener *net.UnixListener
	Conn     net.Conn
	Data     chan interface{}
}

// NewSocket is used by the server command to create a new socket for communication between server and client
func NewSocket(username string) *Socket {
	if err := os.Remove(SocketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Warnf("Could not remove existing socket path: %s (%v)", SocketPath, err)
	}
	addr, err := net.ResolveUnixAddr("unixpacket", SocketPath)
	checkFatal(err)
	listener := listenOnSocket(addr, username)
	conn := acceptOnSocket(listener)
	return &Socket{
		addr:     addr,
		listener: listener,
		Conn:     conn,
		Data:     make(chan interface{}),
	}
}

func listenOnSocket(addr *net.UnixAddr, username string) *net.UnixListener {
	listener, err := net.ListenUnix("unixpacket", addr)
	checkFatal(err)
	u, err := user.Lookup(username)
	checkFatal(err)
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)
	if err := os.Chown(SocketPath, uid, gid); err != nil {
		log.Fatalf("Unable to change ownership on socket file %s to %s:%s : %s", SocketPath, 0, gid, err)
	}
	return listener
}

func acceptOnSocket(listener *net.UnixListener) net.Conn {
	conn, err := listener.Accept()
	checkFatal(err)
	return conn
}

// ConnectSocket is used by the client to connect to the socket the server created for two-way communication
func ConnectSocket() *Socket {
	conn, err := net.Dial("unixpacket", SocketPath)
	checkFatal(err)
	return &Socket{
		Conn: conn,
		Data: make(chan interface{}),
	}
}

func (s *Socket) socketRecv() {
	var rawMsg Msg
	dec := gob.NewDecoder(s.Conn)
	if err := dec.Decode(&rawMsg); err == nil {
		switch {
		case rawMsg.StateMsg != nil:
			s.Data <- rawMsg.StateMsg
		case rawMsg.WordMsg != nil:
			s.Data <- rawMsg.WordMsg
		default:
			log.Errorf("Decoded but unhandled data received: %v", rawMsg)
		}
	}
}

// RecvData handles recieving data on the connection, decoding the message and passing the decoded data to the Data channel (for external processing)
func (s *Socket) RecvData() {
	for {
		s.socketRecv()
	}
}

func (s *Socket) socketSend(msgData interface{}) {
	tryMessage := func() error {
		enc := gob.NewEncoder(s.Conn)
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

// SendState sends a message to the socket of type State
func (s *Socket) SendState(state State) {
	msg := &StateMsg{
		State: state,
	}
	s.socketSend(msg)
}

// SendState sends a message to the socket of type Word
func (s *Socket) SendWord(w string, c string, p rune) {
	msg := &WordMsg{
		Word:       w,
		Correction: c,
		Punct:      p,
	}
	s.socketSend(msg)
}

func checkFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
