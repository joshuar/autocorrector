package control

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"strconv"

	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/nacl/box"
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

type Packet struct {
	EncryptedSize int
	EncryptedData []byte
}

type Socket struct {
	addr      *net.UnixAddr
	Listener  *net.UnixListener
	sharedKey [32]byte
	Conn      net.Conn
	Data      chan interface{}
	Done      chan bool
}

// NewSocket is used by the server command to create a new socket for communication between server and client
func NewSocket(username string) *Socket {
	if err := os.Remove(SocketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Warnf("Could not remove existing socket path: %s (%v)", SocketPath, err)
	}
	addr, err := net.ResolveUnixAddr("unixpacket", SocketPath)
	checkFatal(err)
	log.Debug("Creating socket and waiting for client connection...")
	listener := listenOnSocket(addr, username)
	conn := AcceptOnSocket(listener)
	s := &Socket{
		addr:     addr,
		Listener: listener,
		Conn:     conn,
		Data:     make(chan interface{}),
		Done:     make(chan bool),
	}
	log.Debug("Socket created.")
	s.performHandshake()

	return s
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

func AcceptOnSocket(listener *net.UnixListener) net.Conn {
	conn, err := listener.Accept()
	checkFatal(err)
	return conn
}

// ConnectSocket is used by the client to connect to the socket the server created for two-way communication
func ConnectSocket() *Socket {
	var s *Socket
	tryMessage := func() error {
		conn, err := net.Dial("unixpacket", SocketPath)
		if err != nil {
			return err
		}
		s = &Socket{
			Conn: conn,
			Data: make(chan interface{}),
			Done: make(chan bool),
		}
		log.Debug("Socket connected.")
		s.performHandshake()
		return nil
	}
	err := backoff.Retry(tryMessage, backoff.NewExponentialBackOff())
	if err != nil {
		log.Errorf("Problem with connection backoff", err)
	}
	return s
}

// RecvData handles recieving data on the connection, decoding the message and passing the decoded data to the Data channel (for external processing)
func (s *Socket) RecvData() {
	for {
		var packet Packet
		gobDec := gob.NewDecoder(s.Conn)
		if err := gobDec.Decode(&packet); err != nil {
			log.Errorf("Read error: %s", err)
			if err == io.EOF {
				log.Debug("Sending done channel to indicate restart needed")
				s.Done <- true
				break
			}
		}
		var nonce [24]byte
		copy(nonce[:], packet.EncryptedData[:24])
		encrypted := packet.EncryptedData[24:packet.EncryptedSize]

		decryptedMsg, ok := box.OpenAfterPrecomputation(nil, encrypted, &nonce, &s.sharedKey)
		if !ok {
			log.Warn("Failed to decrypt packet")
		}
		buf := bytes.NewBuffer(decryptedMsg)
		var msg Msg

		dec := gob.NewDecoder(buf)
		if err := dec.Decode(&msg); err == nil {
			switch {
			case msg.StateMsg != nil:
				s.Data <- msg.StateMsg
			case msg.WordMsg != nil:
				s.Data <- msg.WordMsg
			default:
				log.Warnf("Decoded but unhandled data received: %v", msg)
			}
		}
	}
}

func (s *Socket) sendEncrypted(msgData interface{}) {
	tryMessage := func() error {
		var msgBuffer bytes.Buffer
		gobMsg := gob.NewEncoder(&msgBuffer)
		switch t := msgData.(type) {
		case *StateMsg:
			if err := gobMsg.Encode(&Msg{StateMsg: t}); err != nil {
				log.Errorf("Error encoding message: %s", err)
				return err
			}
		case *WordMsg:
			if err := gobMsg.Encode(&Msg{WordMsg: t}); err != nil {
				log.Errorf("Error encoding message: %s", err)
				return err
			}
		default:
			return fmt.Errorf("unknown data to encode: %v", t)
		}
		var nonce [24]byte
		if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
			return err
		}
		e := box.SealAfterPrecomputation(nonce[:], msgBuffer.Bytes(), &nonce, &s.sharedKey)
		packet := &Packet{
			EncryptedSize: len(e),
			EncryptedData: e,
		}
		gobEnc := gob.NewEncoder(s.Conn)
		if err := gobEnc.Encode(&packet); err != nil {
			log.Errorf("Write error: %s", err)
			return err
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
	s.sendEncrypted(msg)
}

// SendState sends a message to the socket of type Word
func (s *Socket) SendWord(w string, c string, p rune) {
	msg := &WordMsg{
		Word:       w,
		Correction: c,
		Punct:      p,
	}
	s.sendEncrypted(msg)
}

func (s *Socket) performHandshake() {
	log.Debug("Performing handshake...")

	var peerKey [32]byte

	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	checkFatal(err)

	// Deliver the public key
	s.Conn.Write(publicKey[:])

	// Receive the peer key
	peerKeyArray := make([]byte, 32)
	s.Conn.Read(peerKeyArray)
	copy(peerKey[:], peerKeyArray)

	box.Precompute(&s.sharedKey, &peerKey, privateKey)
	log.Debug("Handshake complete...")
}

func checkFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
