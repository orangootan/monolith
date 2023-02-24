package main

import (
	"github.com/orangootan/monolith/pkg/monolith"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	name             = "default"
	announceEndPoint = "127.0.0.1:3001"
	requestEndPoint  = "127.0.0.1:3002"
)

func main() {
	dispatcher := runDispatcher()
	waitForSignalAndShutdown(dispatcher)
}

func runDispatcher() *monolith.Dispatcher {
	d := monolith.NewDispatcher(name)
	err := d.ListenAnnounces(announceEndPoint)
	if err != nil {
		log.Fatal(err)
	}
	err = d.Serve(requestEndPoint)
	if err != nil {
		log.Fatal(err)
	}
	return &d
}

func waitForSignalAndShutdown(d *monolith.Dispatcher) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	log.Println("got", sig)
	d.Shutdown()
}
