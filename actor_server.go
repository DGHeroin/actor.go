package actor

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type ActorServer struct {
	sys      *ActorSystem
	listener net.Listener
	conns    sync.WaitGroup

	maxWorkers int
	taskQueue  chan func()
}

func NewActorServer(sys *ActorSystem, maxWorkers int) *ActorServer {
	if maxWorkers <= 0 {
		maxWorkers = 100 // 默认值
	}

	server := &ActorServer{
		sys:        sys,
		maxWorkers: maxWorkers,
		taskQueue:  make(chan func(), maxWorkers),
	}

	for i := 0; i < maxWorkers; i++ {
		go func() {
			for task := range server.taskQueue {
				task()
			}
		}()
	}

	return server
}

func (s *ActorServer) Serve(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	s.listener = listener

	go s.acceptLoop()
	return nil
}
func (s *ActorServer) ServeListener(listener net.Listener) error {
	s.listener = listener

	go s.acceptLoop()
	return nil
}

func (s *ActorServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		s.conns.Add(1)
		go func(c net.Conn) {
			defer s.conns.Done()
			defer c.Close()
			s.handleConnection(c)
		}(conn)
	}
}

func (s *ActorServer) handleConnection(conn net.Conn) {
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)
	var encMu sync.Mutex

	for {
		var msg struct {
			Id      int64       `json:"id"`
			Target  string      `json:"target"`
			Type    MessageType `json:"type"`
			Payload interface{} `json:"payload"`
		}

		if err := dec.Decode(&msg); err != nil {
			return
		}

		switch msg.Type {
		case MessageTypeRequest:
			select {
			case s.taskQueue <- func() {
				respCh := make(chan Response, 1)
				defer close(respCh)

				err := s.sys.SendMessage(msg.Target, Message{
					Payload: msg.Payload,
					ReplyTo: respCh,
				})

				encMu.Lock()
				if err != nil {
					enc.Encode(struct {
						Id    int64  `json:"id"`
						Error string `json:"error"`
					}{
						Id:    msg.Id,
						Error: err.Error(),
					})
					encMu.Unlock()
					return
				}

				resp := <-respCh
				enc.Encode(struct {
					Id      int64       `json:"id"`
					Error   string      `json:"error"`
					Payload interface{} `json:"payload"`
				}{
					Id:      msg.Id,
					Error:   resp.Error,
					Payload: resp.Data,
				})
				encMu.Unlock()
			}:
			default:
				encMu.Lock()
				enc.Encode(struct {
					Id    int64  `json:"id"`
					Error string `json:"error"`
				}{
					Id:    msg.Id,
					Error: "server is busy",
				})
				encMu.Unlock()
			}

		case MessageTypeTell:
			s.sys.SendMessage(msg.Target, Message{
				Payload: msg.Payload,
			})
		}
	}
}

func (s *ActorServer) Stop() error {
	close(s.taskQueue)
	if s.listener != nil {
		s.listener.Close()
	}
	s.conns.Wait()
	return nil
}
