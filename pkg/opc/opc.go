package opc

import (
	"io"
	"log"
	"math"
	"net"
)

// ARGB represents an RGB value with alpha
type ARGB uint32

// Red gets the red component's value
func (a ARGB) Red() byte {
	return byte(a >> 24 & 0xFF)
}

// Green gets the red component's value
func (a ARGB) Green() byte {
	return byte(a >> 16 & 0xFF)
}

// Blue gets the red component's value
func (a ARGB) Blue() byte {
	return byte(a >> 8 & 0xFF)
}

// ARGBColor gets a new aRGB value from its component values
func ARGBColor(a, r, g, b byte) ARGB {
	return ARGB(a | r | g | b)
}

// OPC manages connections to an Open Pixel Control server
type OPC struct {
	Host            string
	Port            int
	Width           int
	Height          int
	PixelLocations  []int
	PacketData      []byte
	FirmwareConfig  byte
	ColorCorrection string
	Output          io.Writer
	Pending         io.Writer
	Connection      net.Conn
	Pixels          []ARGB
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

// Draw maps the pixel array to color values and sends it to the OPC server
func (o *OPC) Draw() error {
	if o.PixelLocations == nil || o.Connection == nil || o.Output == nil {
		return nil
	}

	numPixels := len(o.PixelLocations)
	ledAddress := 4
	o.SetPixelCount(numPixels)

	for i := 0; i < numPixels; i++ {
		pixelLocation := o.PixelLocations[i]
		pixel := o.Pixels[pixelLocation]
		pd := o.PacketData
		pd[ledAddress] = byte(pixel >> 16)
		pd[ledAddress+1] = byte(pixel >> 8)
		pd[ledAddress+2] = byte(pixel)
		ledAddress += 3
	}

	err := o.WritePixels()
	if err != nil {
		return err
	}

	return nil
}

// SetPixelCount sets the total number of pixels
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

// SetPixel sets the color value of a pixel
func (o *OPC) SetPixel(number int, c ARGB) {
	offset := 4 + number*3
	if o.PacketData == nil || len(o.PacketData) < offset+3 {
		o.SetPixelCount(number + 1)
	}

	pd := o.PacketData
	pd[offset] = byte(c >> 16)
	pd[offset+1] = byte(c >> 8)
	pd[offset+2] = byte(c)
}

// WritePixels writes the current buffer of pixel values to the OPC server
func (o *OPC) WritePixels() error {
	if o.PacketData == nil || len(o.PacketData) == 0 || o.Output == nil {
		return nil
	}

	_, err := o.Output.Write(o.PacketData)
	return err
}

// Close disconnects from the OPC server
func (o *OPC) Close() {
	if o.Output != nil {
		log.Println("Disconnecting from OPC server")
		o.Output = nil
	}
	if o.Connection != nil {
		_ = o.Connection.Close()
	}
	o.Pending = nil
}
