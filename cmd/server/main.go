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
	flag.IntVar(&opcChannel, "channel", opc.BROADCAST_CHANNEL, "OPC channel to send to")
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

	for ch := 1; ch < 64; ch++ {
		m := opc.NewMessage(uint8(ch))
		for i := 0; i < 64; i++ {
			m.SetPixelColor(i, 255, 255, 255)
		}
		err = c.Send(m)
		if err != nil {
			log.Fatalf("FATAL: failed to send message: %s", err)
		}
		log.Printf("Sent message to OPC channel %d", ch)
	}
}
