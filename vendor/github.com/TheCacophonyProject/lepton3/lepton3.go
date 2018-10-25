// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package lepton3

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	tomb "gopkg.in/tomb.v2"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/devices/lepton/cci"
)

const (
	// Video Over SPI packets
	vospiHeaderSize = 4 // 2 byte ID, 2 byte CRC
	vospiDataSize   = 160
	vospiPacketSize = vospiHeaderSize + vospiDataSize

	//
	// Packets, segments and frames
	//

	// FrameCols is the X resolution of the Lepton 3 camera.
	FrameCols = 160

	// FrameRows is the Y resolution of the Lepton 3 camera.
	FrameRows = 120

	packetsPerSegment = 60
	segmentsPerFrame  = 4
	packetsPerFrame   = segmentsPerFrame * packetsPerSegment
	colsPerPacket     = FrameCols / 2
	segmentPacketNum  = 20
	maxPacketNum      = 59

	// SPI transfer
	packetsPerRead     = 128
	transferSize       = vospiPacketSize * packetsPerRead
	packetChSize       = 512
	maxPacketsPerFrame = 1500 // including discards and then rounded up somewhat

	// Packet bitmasks
	packetHeaderDiscard = 0x0F
	packetNumMask       = 0x0FFF

	// The maximum time a single frame read is allowed to take
	// (including resync attempts)
	frameTimeout = 10 * time.Second
)

// New returns a new Lepton3 instance.
func New(spiSpeed int64) *Lepton3 {
	// The ring buffer is used to avoid memory allocations for SPI
	// transfers. We aim to have it big enough to handle all the SPI
	// transfers for at least a 3 frames.
	ringChunks := 3 * int(math.Ceil(float64(maxPacketsPerFrame)/float64(packetsPerRead)))
	return &Lepton3{
		spiSpeed:     spiSpeed,
		ring:         newRing(ringChunks, transferSize),
		frameBuilder: newFrameBuilder(),
		log:          func(string) {},
	}
}

// Lepton3 manages a connection to an FLIR Lepton 3 camera. It is not
// goroutine safe.
type Lepton3 struct {
	spiSpeed     int64
	spiPort      spi.PortCloser
	spiConn      spi.Conn
	packetCh     chan []byte
	tomb         *tomb.Tomb
	ring         *ring
	frameBuilder *frameBuilder
	log          func(string)
}

func (d *Lepton3) SetLogFunc(log func(string)) {
	d.log = log
}

// SetRadiometry enables or disables radiometry mode. If enabled, the
// camera will attempt to automatically compensate for ambient
// temperature changes.
func (d *Lepton3) SetRadiometry(enable bool) error {
	cciDev, err := openCCI()
	if err != nil {
		return err
	}
	defer cciDev.Close()

	if err := cciDev.SetRadiometry(enable); err != nil {
		return fmt.Errorf("SetRadiometry: %v", err)
	}
	return nil
}

// RunFFC forces the camera to run a Flat Field Correction
// recalibration.
func (d *Lepton3) RunFFC() error {
	cciDev, err := openCCI()
	if err != nil {
		return err
	}
	defer cciDev.Close()
	return cciDev.RunFFC()
}

// Open initialises the SPI connection and starts streaming packets
// from the camera.
func (d *Lepton3) Open() error {
	spiPort, err := spireg.Open("")
	if err != nil {
		return err
	}
	spiConn, err := spiPort.Connect(d.spiSpeed, spi.Mode3, 8)
	if err != nil {
		spiPort.Close()
		return err
	}

	d.spiPort = spiPort
	d.spiConn = spiConn

	return d.startStream()
}

// Close stops streaming of packets from the camera and closes the SPI
// device connection. It must only be called if streaming was started
// with Open().
func (d *Lepton3) Close() {
	d.stopStream()
	if d.spiPort != nil {
		d.spiPort.Close()
	}
	d.spiConn = nil
}

// NextFrame returns the next frame from the camera into the RawFrame
// provided.
//
// The output RawFrame is provided (rather than being created by
// NextFrame) to minimise memory allocations.
//
// NextFrame should only be called after a successful call to
// Open(). Although there is some internal buffering of camera
// packets, NextFrame must be called frequently enough to ensure
// frames are not lost.
func (d *Lepton3) NextFrame(outFrame *RawFrame) error {
	timeout := time.After(frameTimeout)
	d.frameBuilder.reset()

	var packet []byte
	for {
		select {
		case packet = <-d.packetCh:
		case <-d.tomb.Dying():
			if err := d.tomb.Err(); err != nil {
				return fmt.Errorf("streaming failed: %v", err)
			}
			return nil
		case <-timeout:
			return errors.New("frame timeout")
		}

		packetNum, err := validatePacket(packet)
		if err != nil {
			if err := d.resync(err); err != nil {
				return err
			}
			continue
		} else if packetNum < 0 {
			continue
		}

		complete, err := d.frameBuilder.nextPacket(packetNum, packet)
		if err != nil {
			if err := d.resync(err); err != nil {
				return err
			}
		} else if complete {
			d.frameBuilder.output(outFrame)
			return nil
		}
	}
}

// Snapshot is convenience method for capturing a single frame. It
// should *not* be called if streaming is already active.
func (d *Lepton3) Snapshot() (*RawFrame, error) {
	if err := d.Open(); err != nil {
		return nil, err
	}
	defer d.Close()
	frame := new(RawFrame)
	if err := d.NextFrame(frame); err != nil {
		return nil, err
	}
	return frame, nil
}

func (d *Lepton3) resync(reason error) error {
	d.log(fmt.Sprintf("resync! %v", reason))
	d.Close()
	d.frameBuilder.reset()
	time.Sleep(300 * time.Millisecond)
	return d.Open()
}

func (d *Lepton3) startStream() error {
	if d.tomb != nil {
		return errors.New("streaming already active")
	}
	d.tomb = new(tomb.Tomb)
	d.packetCh = make(chan []byte, packetChSize)
	d.tomb.Go(func() error {
		for {
			rx := d.ring.next()
			if err := d.spiConn.Tx(nil, rx); err != nil {
				return err
			}
			for i := 0; i < len(rx); i += vospiPacketSize {
				if rx[i]&packetHeaderDiscard == packetHeaderDiscard {
					// No point sending discard packets onwards.
					// This makes a big difference to CPU utilisation.
					continue
				}
				select {
				case <-d.tomb.Dying():
					return tomb.ErrDying
				case d.packetCh <- rx[i : i+vospiPacketSize]:
				}
			}
		}
	})
	return nil
}

func (d *Lepton3) stopStream() {
	if d.tomb != nil {
		d.tomb.Kill(nil)
		d.tomb.Wait()
		d.tomb = nil
	}
}

func validatePacket(packet []byte) (int, error) {
	header := binary.BigEndian.Uint16(packet)
	if header&0x8000 == 0x8000 {
		return -1, errors.New("first bit set on header")
	}

	packetNum := int(header & packetNumMask)
	if packetNum > 60 {
		return -1, errors.New("invalid packet number")
	}

	// XXX might not necessary with CRC check
	if packetNum == 0 && packet[2] == 0 && packet[3] == 0 {
		return -1, nil
	}

	// XXX CRC checks

	return packetNum, nil
}

func openCCI() (*closingCCIDev, error) {
	i2cBus, err := i2creg.Open("")
	if err != nil {
		return nil, err
	}

	cciDev, err := cci.New(i2cBus)
	if err != nil {
		i2cBus.Close()
		return nil, fmt.Errorf("cci.New: %v", err)
	}
	return &closingCCIDev{
		Dev:    cciDev,
		Closer: i2cBus,
	}, nil
}

type closingCCIDev struct {
	*cci.Dev
	io.Closer
}
