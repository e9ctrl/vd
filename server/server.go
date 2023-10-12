package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	CONN_TYPE = "tcp"
	BUF_SIZE  = 4096
)

type Handler interface {
	Handle([]byte) []byte
}

// Server struct
type Server struct {
	wg         sync.WaitGroup
	listener   net.Listener
	shutdown   chan struct{}
	connection chan net.Conn
	d          Handler
}

// Create a new server with given handler and address
func New(device Handler, address string) (*Server, error) {
	listener, err := net.Listen(CONN_TYPE, address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on address %s: %w", address, err)
	}

	return &Server{
		listener:   listener,
		shutdown:   make(chan struct{}),
		connection: make(chan net.Conn),
		d:          device,
	}, nil
}

func (s *Server) acceptConnections() {
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdown:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				continue
			}
			s.connection <- conn
		}
	}
}

func (s *Server) handleConnections() {
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdown:
			return
		case conn := <-s.connection:
			go s.handleConnection(conn)
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, BUF_SIZE)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("error reading connection", err.Error())
			}
			break
		}

		response := s.d.Handle(buffer[:n])
		_, writeErr := conn.Write(response)
		if writeErr != nil {
			fmt.Println("error writing response", writeErr.Error())
			break
		}
	}
}

// Start server
func (s *Server) Start() {
	s.wg.Add(2)
	go s.acceptConnections()
	go s.handleConnections()
}

// Stop server
func (s *Server) Stop() {
	close(s.shutdown)
	s.listener.Close()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(time.Second):
		fmt.Println("Timed out waiting for connections to finish.")
		return
	}
}
