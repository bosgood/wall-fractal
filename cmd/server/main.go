package main

import (
	"flag"
	"log"
	"time"

	"github.com/bosgood/wall-fractal/pkg/opc"
)

var (
	addr string
)

func parseargs() {
	flag.StringVar(&addr, "addr", "127.0.0.1:7890", "Address and port of the OPC server to connect to")
	flag.Parse()
}

func main() {
	log.Println("wall-fractal starting")
	parseargs()

	width := 800
	height := 200
	c := opc.New(addr, width, height, 0*time.Second)

	x := float64(width) / 2
	c.LEDStrip(0, 64, x, float64(height)/2-30, float64(width)/70.0, 0, false)
	c.LEDStrip(65, 64, x, float64(height)/2+0, float64(width)/70.0, 0, false)
	c.LEDStrip(129, 64, x, float64(height)/2+30, float64(width)/70.0, 0, false)
	c.LEDStrip(193, 64, x, float64(height)/2+60, float64(width)/70.0, 0, false)

	numPixels := width * height
	pixels := make([]opc.Color, numPixels)
	for i := 0; i < numPixels; i++ {
		pixels[i] = opc.ColorFromARGB(255, 255, 255, 255)
	}
	c.Refresh(pixels)
	go c.Run()
	select {}
}
