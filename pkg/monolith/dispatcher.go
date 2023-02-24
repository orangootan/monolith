package monolith

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Dispatcher struct {
	name      string
	services  SyncMap[string, string]
	listeners []*net.TCPListener
	wgs       []*sync.WaitGroup
	logger    *log.Logger
}

func NewDispatcher(name string) Dispatcher {
	return Dispatcher{
		name:     name,
		services: NewSyncMap[string, string](),
		logger:   log.Default(),
	}
}

func (d *Dispatcher) logf(format string, v ...any) {
	if d.logger == nil {
		return
	}
	message := fmt.Sprintf(format, v...)
	d.logger.Printf("Dispatcher '%v': %v\n", d.name, message)
}

func (d *Dispatcher) log(v ...any) {
	if d.logger == nil {
		return
	}
	message := fmt.Sprint(v...)
	d.logger.Printf("Dispatcher '%v': %v\n", d.name, message)
}

func (d *Dispatcher) SetLogger(logger *log.Logger) {
	d.logger = logger
}

func (d *Dispatcher) Name() string {
	return d.name
}

func (d *Dispatcher) Stop() {
	d.log("stopping...")
	for _, listener := range d.listeners {
		err := listener.Close()
		if err != nil {
			d.log(err)
		}
	}
}

func (d *Dispatcher) Wait() {
	for _, wg := range d.wgs {
		wg.Wait()
	}
	d.log("graceful shutdown complete.")
}

func (d *Dispatcher) Shutdown() {
	d.Stop()
	d.Wait()
}

func (d *Dispatcher) ListenAnnounces(endPoint string) (err error) {
	local, err := net.ResolveTCPAddr("tcp", endPoint)
	if err != nil {
		return
	}
	listener, err := net.ListenTCP("tcp", local)
	if err != nil {
		return
	}
	d.log("started listening service announces on ", endPoint)
	d.listeners = append(d.listeners, listener)
	var wg sync.WaitGroup
	d.wgs = append(d.wgs, &wg)
	go func() {
		defer d.log("stopped listening service announces on ", endPoint)
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				d.log(err)
				return
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := d.addServices(conn)
				if err != nil && err != io.EOF {
					d.log(err)
				}
			}()
		}
	}()
	return
}

func (d *Dispatcher) addServices(conn net.Conn) (err error) {
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				d.log(closeErr)
			}
		}
	}()
	remote := conn.RemoteAddr().String()
	d.log("server connected from address ", remote)
	decoder := gob.NewDecoder(conn)
	for {
		var service string
		err = decoder.Decode(&service)
		if err != nil {
			return
		}
		d.services.put(service, remote)
		d.logf("server %v announced service '%v'", remote, service)
	}
}

func (d *Dispatcher) Serve(endPoint string) (err error) {
	local, err := net.ResolveTCPAddr("tcp", endPoint)
	if err != nil {
		return
	}
	listener, err := net.ListenTCP("tcp", local)
	if err != nil {
		return
	}
	d.log("started listening client connections on ", endPoint)
	d.listeners = append(d.listeners, listener)
	var wg sync.WaitGroup
	d.wgs = append(d.wgs, &wg)
	go func() {
		defer d.log("stopped listening client connections on ", endPoint)
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				d.log(err)
				return
			}
			remote := conn.RemoteAddr().String()
			d.log("client connected from address ", remote)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := d.respond(conn)
				if err != nil && err != io.EOF {
					d.log(err)
				} else {
					d.log("client ", remote, " disconnected")
				}
			}()
		}
	}()
	return
}

func (d *Dispatcher) respond(conn net.Conn) (err error) {
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				d.log(closeErr)
			}
		}
	}()
	remote := conn.RemoteAddr().String()
	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)
	for {
		var service string
		err = decoder.Decode(&service)
		if err != nil {
			return
		}
		d.logf("client %v requested service '%v'", remote, service)
		endPoint, _ := d.services.get(service)
		err = encoder.Encode(endPoint)
		if err != nil {
			return
		}
		d.logf("responded to client %v: service '%v' has address %v", remote, service, endPoint)
	}
}
