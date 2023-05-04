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
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/nacl/box"
)

const (
	SocketPath       = "/tmp/autocorrector.sock"
	Resume     State = 0x01
	Pause      State = 0x00
)

type State byte

type StateMsg struct {
	State
}

type WordMsg struct {
	Word, Correction string
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

// CreateServer is used by the server command to create a new socket for communication between server and client
func CreateServer(username string) *Socket {
	defer func() {
		if r := recover(); r != nil {
			log.Debug().Caller().
				Msgf("Error in NewSocket: %v", r)
		}
	}()
	if err := os.Remove(SocketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Debug().Caller().Err(err).
			Msgf("Could not remove existing socket path: %s.", SocketPath)
	}
	addr, err := net.ResolveUnixAddr("unixpacket", SocketPath)
	checkFatal(err)
	log.Debug().Caller().
		Msg("Creating socket and waiting for client connection...")
	listener := listenOnSocket(addr, username)
	conn := acceptOnSocket(listener)
	s := &Socket{
		addr:     addr,
		Listener: listener,
		Conn:     conn,
		Data:     make(chan interface{}),
		Done:     make(chan bool),
	}
	log.Debug().Caller().
		Msg("Socket created.")
	s.performHandshake()
	go s.recvData()
	return s
}

// ConnectSocket is used by the client to connect to the socket the server created for two-way communication
func CreateClient() *Socket {
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
		log.Debug().Caller().
			Msg("Socket connected.")
		s.performHandshake()
		return nil
	}
	err := backoff.Retry(tryMessage, backoff.NewExponentialBackOff())
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Problem with connection backoff.")
	}
	go s.recvData()
	s.ResumeServer()
	return s
}

func listenOnSocket(addr *net.UnixAddr, username string) *net.UnixListener {
	defer func() {
		if r := recover(); r != nil {
			log.Debug().Caller().
				Msgf("Error in listenOnSocket: %v", r)
		}
	}()
	listener, err := net.ListenUnix("unixpacket", addr)
	checkFatal(err)
	u, err := user.Lookup(username)
	checkFatal(err)
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)
	if err := os.Chown(SocketPath, uid, gid); err != nil {
		log.Debug().Caller().Err(err).
			Msgf("Unable to change ownership on socket file %s to %s:%s : %s", SocketPath, 0, gid, err)
	}
	return listener
}

func acceptOnSocket(listener *net.UnixListener) net.Conn {
	defer func() {
		if r := recover(); r != nil {
			log.Debug().Caller().
				Msgf("Error in AcceptOnSocket: %v", r)
		}
	}()
	conn, err := listener.Accept()
	checkFatal(err)
	return conn
}

func (s *Socket) recvData() {
	for {
		var packet Packet
		gobDec := gob.NewDecoder(s.Conn)
		if err := gobDec.Decode(&packet); err != nil {
			log.Debug().Caller().Err(err).Msg("Read error.")
			if err == io.EOF {
				log.Debug().Caller().
					Msg("Sending done channel to indicate restart needed.")
				s.Done <- true
				break
			}
		}
		var nonce [24]byte
		copy(nonce[:], packet.EncryptedData[:24])
		encrypted := packet.EncryptedData[24:packet.EncryptedSize]

		decryptedMsg, ok := box.OpenAfterPrecomputation(nil, encrypted, &nonce, &s.sharedKey)
		if !ok {
			log.Debug().Caller().
				Msg("Failed to decrypt packet")
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
				log.Debug().Caller().
					Msgf("Decoded but unhandled data received: %v", msg)
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
				log.Debug().Caller().Err(err).
					Msg("Error encoding message.")
				return err
			}
		case *WordMsg:
			if err := gobMsg.Encode(&Msg{WordMsg: t}); err != nil {
				log.Debug().Caller().Err(err).
					Msg("Error encoding message.")
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
			log.Debug().Caller().Err(err).
				Msg("Write error.")
			return err
		}
		return nil
	}
	err := backoff.Retry(tryMessage, backoff.NewExponentialBackOff())
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Problem with connection backoff.")
	}
}

// PauseServer sends a state message to indicate the server should pause
// tracking key presses
func (s *Socket) PauseServer() {
	msg := &StateMsg{
		State: Pause,
	}
	s.sendEncrypted(msg)
}

// ResumeServer sends a state message to indicate the server should resume
// tracking key presses
func (s *Socket) ResumeServer() {
	msg := &StateMsg{
		State: Resume,
	}
	s.sendEncrypted(msg)
}

// SendState sends a message to the socket of type Word
func (s *Socket) SendWord(w *WordMsg) {
	s.sendEncrypted(w)
}

func (s *Socket) performHandshake() {
	defer func() {
		if r := recover(); r != nil {
			log.Debug().Caller().
				Msgf("Error in performHandShake: %v", r)
		}
	}()

	log.Debug().Caller().Msg("Performing handshake...")

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
	log.Debug().Caller().Msg("Handshake complete...")
}

func checkFatal(err error) {
	if err != nil {
		panic(err)
	}
}
