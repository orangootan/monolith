package main

import (
	"fmt"
	"github.com/orangootan/monolith/pkg/monolith"
	"log"
)

const (
	name               = "default"
	clientEndPoint     = "127.0.0.1:0"
	dispatcherEndPoint = "127.0.0.1:3002"
)

func main() {
	runClient()
}

func runClient() {
	client, err := monolith.NewClient(name, clientEndPoint, dispatcherEndPoint)
	if err != nil {
		log.Fatal(err)
	}
	math, err := monolith.Get[Math]("1", &client)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(math.Add(1, 2))
	fmt.Println(math.Divide(1, 0))
	fmt.Println(math.Sqrt(4))
}
