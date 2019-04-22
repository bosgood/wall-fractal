package opc

import (
	"io"
	"math"
	"net"
)

// Color represents an RGB value
type Color [3]byte

// Red gets the red component's value
func (c *Color) Red() byte {
	return c[0]
}

// Green gets the red component's value
func (c *Color) Green() byte {
	return c[1]
}

// Blue gets the red component's value
func (c *Color) Blue() byte {
	return c[2]
}

func (c *Color) Int32() uint32 {
	return 0
}

// OPC manages connections to an Open Pixel Control server
type OPC struct {
	Host                string
	Port                int
	Width               int
	Height              int
	PixelLocations      []int
	PacketData          []byte
	FirmwareConfig      byte
	ColorCorrection     string
	EnableShowLocations bool
	Output              io.Writer
	Connection          net.Conn
	Pixels              []Color
}

// New creates a new OPC client
func New(host string, port int, width, height int) *OPC {
	return &OPC{
		Host:   host,
		Port:   port,
		Width:  width,
		Height: height,
	}
}

func (o *OPC) LED(index, x, y int) {
	// For convenience, automatically grow the PixelLocations array
	if o.PixelLocations == nil {
		o.PixelLocations = make([]int, index+1)
	} else if index >= len(o.PixelLocations) {
		pl := o.PixelLocations
		o.PixelLocations = make([]int, index+1)
		for i := range pl {
			o.PixelLocations[i] = pl[i]
		}
	}
	o.PixelLocations[index] = x + o.Width*y
}

func (o *OPC) LEDStrip(index, count int, x, y, spacing, angle float64, reversed bool) {
	s := math.Sin(angle)
	c := math.Cos(angle)
	for i := 0; i < count; i++ {
		var idx int
		if reversed {
			idx = index + count - 1 - i
		} else {
			idx = index + 1
		}
		o.LED(
			idx,
			int(x+float64(i-(count-1)/2.0)*spacing*c+0.5),
			int(y+float64(i-(count-1)/2.0)*spacing*s+0.5),
		)
	}
}

func (o *OPC) Draw() error {
	if o.PixelLocations == nil || o.Connection == nil || o.Output == nil {
		return nil
	}

	numPixels := len(o.PixelLocations)
	ledAddress := 4
	o.SetPixelCount(numPixels)

	for i := 0; i < numPixels; i++ {
		loc := o.PixelLocations[i]
		pixel := o.Pixels[loc]
		o.PacketData[ledAddress] = byte(pixel >> 16)

	}

	return nil
}

func (o *OPC) SetPixelCount(numPixels int) {
	numBytes := 3 * numPixels
	packetLen := 4 + numBytes
	if o.PacketData == nil || len(o.PacketData) != packetLen {
		// Set up our packet buffer
		pd := make([]byte, packetLen)
		pd[0] = byte(0x00)            // Channel
		pd[1] = byte(0x00)            // Command (Set pixel colors)
		pd[2] = byte(numBytes >> 8)   // Length high byte
		pd[3] = byte(numBytes & 0xFF) // Length low byte
		o.PacketData = pd
	}
}
