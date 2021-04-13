package control

import (
	"io"
	"net"
	"os"
	"os/user"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/davecgh/go-spew/spew"
)

type ControlSocket struct {
	name net.Listener
	conn net.Conn
	path string
	user *user.User
}

func newServerSocket(username string) *ControlSocket {
	user, err := user.Lookup(username)
	if err != nil {
		log.Fatal(err)
	}
	socketPath := "/tmp/autocorrector-" + user.Username + "-server.sock"
	uid, _ := strconv.Atoi(user.Uid)
	gid, _ := strconv.Atoi(user.Gid)
	os.Remove(socketPath)
	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Unable to listen on socket file %s: %s", socketPath, err)
	}
	if err := os.Chown(socketPath, uid, gid); err != nil {
		log.Fatalf("Unable to change permissions on socket file %s to %s:%s : %s", socketPath, uid, gid, err)
	}
	return &ControlSocket{
		name: socket,
		path: socketPath,
		user: user,
	}
}

func newClientSocket() *ControlSocket {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	socketPath := "/tmp/autocorrector-" + user.Username + "-client.sock"
	os.Remove(socketPath)
	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Unable to listen on socket file %s: %s", socketPath, err)
	}
	return &ControlSocket{
		name: socket,
		path: socketPath,
		user: user,
	}
}

func (s *ControlSocket) AcceptConnections() {
	for {
		conn, err := s.name.Accept()
		if err != nil {
			log.Fatalf("Error on accept: %s", err)
		}
		s.conn = conn
		go s.handleConn()
	}
}

func (s *ControlSocket) handleConn() {
	log.Debug("Server started and listening...")
	received := make([]byte, 0)
	for {
		buf := make([]byte, 512)
		count, err := s.conn.Read(buf)
		received = append(received, buf[:count]...)
		if err != nil {
			spew.Dump(received)
			if err != io.EOF {
				log.Errorf("Error on read: %s", err)
			}
			break
		}
	}
}
