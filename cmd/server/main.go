package main

import (
	"flag"
	"log"

	"github.com/kellydunn/go-opc"
)

var (
	addr       string
	opcChannel int
)

func parseargs() {
	flag.StringVar(&addr, "addr", "127.0.0.1:7890", "Address and port of the OPC server to connect to")
	flag.IntVar(&opcChannel, "channel", 0, "OPC channel to send to")
	flag.Parse()
}

func main() {
	log.Println("wall-fractal starting")

	parseargs()

	c := opc.NewClient()
	err := c.Connect("tcp", addr)
	if err != nil {
		log.Fatalf("FATAL: failed to connect to OPC server at: %s", addr)
	}

	log.Printf("Connected to server at: %s", addr)

	// Make a message!
	// This creates a message to send on channel 0
	// Or according to the OPC spec, a Broadcast Message.
	m := opc.NewMessage(uint8(opcChannel))

	for i := 0; i < 20; i++ {
		m.SetPixelColor(i, 255, 255, 255)
	}

	// Send the message!
	err = c.Send(m)
	if err != nil {
		log.Fatalf("FATAL: failed to send message: %s", err)
	}
	log.Printf("Sent message to OPC channel %d", opcChannel)

	// The first pixel of all registered devices should be white!
}
