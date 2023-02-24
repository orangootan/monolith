package monolith

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type TypeHandler func(
	id string,
	method string,
	decode func(params any) error,
	encode func(results any) error) error

var typeHandlers = make(map[string]TypeHandler)

func RegisterTypeHandler(name string, handler TypeHandler) {
	typeHandlers[name] = handler
}

type Server struct {
	name      string
	listeners []*net.TCPListener
	wgs       []*sync.WaitGroup
	logger    *log.Logger
}

func NewServer(name string) Server {
	return Server{
		name:   name,
		logger: log.Default(),
	}
}

func (s *Server) Name() string {
	return s.name
}

func (s *Server) Stop() {
	s.log("stopping...")
	for _, listener := range s.listeners {
		err := listener.Close()
		if err != nil {
			s.log(err)
		}
	}
}

func (s *Server) Wait() {
	for _, wg := range s.wgs {
		wg.Wait()
	}
	s.log("graceful shutdown complete.")
}

func (s *Server) Shutdown() {
	s.Stop()
	s.Wait()
}

func (s *Server) logf(format string, v ...any) {
	if s.logger == nil {
		return
	}
	message := fmt.Sprintf(format, v...)
	s.logger.Printf("Server '%v': %v\n", s.name, message)
}

func (s *Server) log(v ...any) {
	if s.logger == nil {
		return
	}
	message := fmt.Sprint(v...)
	s.logger.Printf("Server '%v': %v\n", s.name, message)
}

func (s *Server) Serve(endPoint string) (err error) {
	local, err := net.ResolveTCPAddr("tcp", endPoint)
	if err != nil {
		return
	}
	listener, err := net.ListenTCP("tcp", local)
	if err != nil {
		return
	}
	s.log("started listening connections on ", endPoint)
	s.listeners = append(s.listeners, listener)
	var wg sync.WaitGroup
	s.wgs = append(s.wgs, &wg)
	go func() {
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				s.log(err)
				return
			}
			remote := conn.RemoteAddr().String()
			s.log("client connected from address ", remote)
			wg.Add(1)
			go func(conn net.Conn) {
				defer wg.Done()
				err := s.listen(conn, &wg)
				if err != nil && err != io.EOF {
					s.log(err)
				} else {
					s.log("client ", remote, " disconnected")
				}
			}(conn)
		}
	}()
	return
}

func (s *Server) AnnounceServices(serverEndPoint, announceEndPoint string) (err error) {
	local, err := net.ResolveTCPAddr("tcp", serverEndPoint)
	if err != nil {
		return
	}
	remote, err := net.ResolveTCPAddr("tcp", announceEndPoint)
	if err != nil {
		return
	}
	conn, err := net.DialTCP("tcp", local, remote)
	if err != nil {
		return
	}
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				s.log(closeErr)
			}
		}
	}()
	s.log("connected to dispatcher ", announceEndPoint)
	encoder := gob.NewEncoder(conn)
	for service := range typeHandlers {
		err = encoder.Encode(service)
		if err != nil {
			return
		}
		s.logf("successfully announced service '%v'", service)
	}
	return
}

func (s *Server) listen(conn net.Conn, wg *sync.WaitGroup) (err error) {
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				s.log(closeErr)
			}
		}
	}()
	remote := conn.RemoteAddr().String()
	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)
	for {
		var req request
		err = decoder.Decode(&req)
		if err != nil {
			return
		}
		s.log("received request with ID ", req.ID, " from client ", remote)
		wg.Add(1)
		go func(req request) {
			defer wg.Done()
			res := process(req)
			if res.Err != nil {
				s.log(res.Err)
				res.Err = NewError(res.Err.Error())
			}
			err = encoder.Encode(res)
			if err != nil {
				s.log(err)
			} else {
				s.log("successfully sent response with ID ", res.ID, " to client ", remote)
			}
		}(req)
	}
}

func process(req request) (res response) {
	res.ID = req.ID
	handler, ok := typeHandlers[req.Instance.Type]
	if !ok {
		res.Err = UnregisteredTypeError
		return
	}
	decoder := gob.NewDecoder(bytes.NewBuffer(req.Params))
	decode := func(params any) error {
		return decoder.Decode(params)
	}
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	encode := func(results any) error {
		return encoder.Encode(results)
	}
	res.Err = handler(req.Instance.ID, req.Method, decode, encode)
	res.Results = buffer.Bytes()
	return
}
