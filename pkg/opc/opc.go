package opc

import (
	"log"
	"math"
	"net"
	"time"
)

const (
	// DefaultRefreshRate is the frequency at which the pixel data will be sent
	DefaultRefreshRate = 500 * time.Millisecond
	// DefaultReceiveBufferSize is the number of screens to buffer in memory if the connection is backed up
	DefaultReceiveBufferSize = 25
)

// Color represents an RGB value with alpha
type Color uint32

// Red gets the red component's value
func (a Color) Red() byte {
	return byte(a >> 24 & 0xFF)
}

// Green gets the red component's value
func (a Color) Green() byte {
	return byte(a >> 16 & 0xFF)
}

// Blue gets the red component's value
func (a Color) Blue() byte {
	return byte(a >> 8 & 0xFF)
}

// ColorFromARGB gets a new Color value from its component values
func ColorFromARGB(a, r, g, b byte) Color {
	return Color(a | r | g | b)
}

// OPC manages connections to an Open Pixel Control server
type OPC struct {
	addr              string
	width             int
	height            int
	pixelLocations    []int
	packetData        []byte
	firmwareConfig    byte
	colorCorrection   string
	connection        net.Conn
	pixels            []Color
	stop              chan struct{}
	receive           chan []Color
	refreshRate       time.Duration
	receiveBufferSize int
}

// New creates a new OPC client
func New(addr string, width, height int, refreshRate time.Duration) *OPC {
	var rate time.Duration
	if refreshRate == rate {
		rate = DefaultRefreshRate
	}
	return &OPC{
		addr:              addr,
		width:             width,
		height:            height,
		stop:              make(chan struct{}),
		receive:           make(chan []Color, DefaultReceiveBufferSize),
		refreshRate:       rate,
		receiveBufferSize: DefaultReceiveBufferSize,
	}
}

// Refresh refreshes the pixels displayed
func (o *OPC) Refresh(pixels []Color) {
	// Drop all new updates on the floor once the buffer is full
	if len(o.receive) == o.receiveBufferSize {
		return
	}
	o.receive <- pixels
}

// LED initializes a set of LEDs
// This method is not typically called directly; rather, use one of the
// helpers, e.g., LEDStrip, based on the type of LED hardware you have
func (o *OPC) LED(index, x, y int) {
	// For convenience, automatically grow the PixelLocations array
	if o.pixelLocations == nil {
		o.pixelLocations = make([]int, index+1)
	} else if index >= len(o.pixelLocations) {
		pl := o.pixelLocations
		o.pixelLocations = make([]int, index+1)
		for i := range pl {
			o.pixelLocations[i] = pl[i]
		}
	}
	o.pixelLocations[index] = x + o.width*y
}

// LEDStrip initializes a strip of LEDs
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

// TODO [bosgood] Implement
// func (o *OPC) LEDRing()
// func (o *OPC) LEDGrid()
// func (o *OPC) LEDGrid8x8()
// func (o *OPC) SetDithering(enabled bool)
// func (o *OPC) SetInterpolation(enabled bool)
// func (o *OPC) StatusLEDAuto()
// func (o *OPC) SetStatusLED(on bool)
// func (o *OPC) SetColorCorrection(s string)
// func (o *OPC) SendFirmwareConfigPacket(s string)
// func (o *OPC) SendColorCorrectionPacket(s string)

// draw maps the pixel array to color values and sends it to the OPC server
func (o *OPC) draw() error {
	numPixels := len(o.pixelLocations)
	ledAddress := 4
	o.SetPixelCount(numPixels)

	for i := 0; i < numPixels; i++ {
		pixelLocation := o.pixelLocations[i]
		pixel := o.pixels[pixelLocation]
		pd := o.packetData
		pd[ledAddress] = byte(pixel >> 16)
		pd[ledAddress+1] = byte(pixel >> 8)
		pd[ledAddress+2] = byte(pixel)
		ledAddress += 3
	}

	err := o.writePixels()
	if err != nil {
		return err
	}

	return nil
}

// SetPixelCount sets the total number of pixels
func (o *OPC) SetPixelCount(numPixels int) {
	numBytes := 3 * numPixels
	packetLen := 4 + numBytes
	if o.packetData == nil || len(o.packetData) != packetLen {
		// Set up our packet buffer
		pd := make([]byte, packetLen)
		pd[0] = byte(0x00)            // Channel
		pd[1] = byte(0x00)            // Command (Set pixel colors)
		pd[2] = byte(numBytes >> 8)   // Length high byte
		pd[3] = byte(numBytes & 0xFF) // Length low byte
		o.packetData = pd
	}
}

// SetPixel sets the color value of a pixel
func (o *OPC) SetPixel(number int, c Color) {
	offset := 4 + number*3
	if o.packetData == nil || len(o.packetData) < offset+3 {
		o.SetPixelCount(number + 1)
	}

	pd := o.packetData
	pd[offset] = byte(c >> 16)
	pd[offset+1] = byte(c >> 8)
	pd[offset+2] = byte(c)
}

// writePixels writes the current buffer of pixel values to the OPC server
func (o *OPC) writePixels() error {
	if o.packetData == nil || len(o.packetData) == 0 {
		return nil
	}

	_, err := o.connection.Write(o.packetData)
	return err
}

// Stop disconnects from the OPC server
func (o *OPC) Stop() {
	o.stop <- struct{}{}
}

func (o *OPC) connect() error {
	if o.connection == nil {
		conn, err := net.Dial("tcp", o.addr)
		if err != nil {
			log.Printf("Error connecting to OPC server: %v", err)
			return err
		}
		o.connection = conn
	}
	return nil
}

// Run starts the reconnection message loop
func (o *OPC) Run() {
	t := time.NewTicker(o.refreshRate)
	defer func() {
		if o.connection != nil {
			log.Println("Closing connection to OPC server")
			_ = o.connection.Close()
		}
	}()

	for {
		select {
		case <-o.stop:
			break
		case p := <-o.receive:
			o.pixels = p
		case <-t.C:
			err := o.connect()
			if err != nil {
				log.Printf("Error connecting to OPC server: %v\n", err)
				continue
			}
			err = o.draw()
			if err != nil {
				log.Printf("Error sending pixel data to server: %v\n", err)
				continue
			}
		}
	}
}
