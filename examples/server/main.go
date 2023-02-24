package main

import (
	"github.com/orangootan/monolith/pkg/monolith"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	name                       = "default"
	serverEndPoint             = "127.0.0.1:3000"
	dispatcherAnnounceEndPoint = "127.0.0.1:3001"
)

func main() {
	server := runServer()
	waitForSignalAndShutdown(server)
}

func runServer() *monolith.Server {
	s := monolith.NewServer(name)
	err := s.AnnounceServices(serverEndPoint, dispatcherAnnounceEndPoint)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Serve(serverEndPoint)
	if err != nil {
		log.Fatal(err)
	}
	return &s
}

func waitForSignalAndShutdown(s *monolith.Server) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	log.Println("got", sig)
	s.Shutdown()
}
