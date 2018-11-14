// thermal-recorder - record thermal video footage of warm moving objects
//  Copyright (C) 2018, The Cacophony Project
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"github.com/TheCacophonyProject/lepton3"
	arg "github.com/alexflint/go-arg"
	"github.com/coreos/go-systemd/daemon"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

const (
	framesHz = 9 // approx

	frameLogIntervalFirstMin = 15 * framesHz
	frameLogInterval         = 60 * 5 * framesHz

	framesPerSdNotify = 5 * framesHz
)

var version = "<not set>"

type Args struct {
	ConfigFile string `arg:"-c,--config" help:"path to configuration file"`
	Quick      bool   `arg:"-q,--quick" help:"don't cycle camera power on startup"`
	Timestamps bool   `arg:"-t,--timestamps" help:"include timestamps in log output"`
}

func (Args) Version() string {
	return version
}

func procArgs() Args {
	var args Args
	args.ConfigFile = "/etc/leptond.yaml"
	arg.MustParse(&args)
	return args
}

type nextFrameErr struct {
	cause error
}

func (e *nextFrameErr) Error() string {
	return e.cause.Error()
}

func main() {
	err := runMain()
	if err != nil {
		log.Fatal(err)
	}
}

func runMain() error {
	args := procArgs()
	if !args.Timestamps {
		log.SetFlags(0) // Removes default timestamp flag
	}

	log.Printf("version: %s", version)
	conf, err := ParseConfigFile(args.ConfigFile)
	if err != nil {
		return err
	}
	logConfig(conf)

	log.Print("dialing frame output socket")
	conn, err := net.DialUnix("unixpacket", nil, &net.UnixAddr{
		Net:  "unixgram",
		Name: conf.FrameOutput,
	})
	if err != nil {
		return errors.New("error: connecting to frame output socket failed")
	}
	defer conn.Close()

	log.Print("host initialisation")
	if _, err := host.Init(); err != nil {
		return err
	}

	if !args.Quick {
		if err := cycleCameraPower(conf.PowerPin); err != nil {
			return err
		}
	}

	var camera *lepton3.Lepton3
	defer func() {
		if camera != nil {
			camera.Close()
		}
	}()
	for {
		camera, err = lepton3.New(conf.SPISpeed)
		if err != nil {
			return err
		}
		camera.SetLogFunc(func(t string) { log.Printf(t) })

		log.Print("enabling radiometry")
		if err := camera.SetRadiometry(true); err != nil {
			return err
		}

		log.Print("opening camera")
		if err := camera.Open(); err != nil {
			return err
		}

		err := runCamera(conf, camera, conn)
		if err != nil {
			if _, isNextFrameErr := err.(*nextFrameErr); !isNextFrameErr {
				return err
			}
			log.Printf("recording error: %v", err)
		}

		log.Print("closing camera")
		camera.Close()

		err = cycleCameraPower(conf.PowerPin)
		if err != nil {
			return err
		}
	}
}

func runCamera(conf *Config, camera *lepton3.Lepton3, conn *net.UnixConn) error {
	conn.SetWriteBuffer(lepton3.FrameCols * lepton3.FrameRows * 2 * 20)

	log.Print("reading frames")
	frame := new(lepton3.RawFrame)
	notifyCount := 0
	for {
		if err := camera.NextFrame(frame); err != nil {
			return &nextFrameErr{err}
		}

		if notifyCount++; notifyCount >= framesPerSdNotify {
			daemon.SdNotify(false, "WATCHDOG=1")
			notifyCount = 0
		}

		if _, err := conn.Write(frame[:]); err != nil {
			return err
		}
	}
}

func logConfig(conf *Config) {
	log.Printf("SPI speed: %d", conf.SPISpeed)
	log.Printf("power pin: %s", conf.PowerPin)
	log.Printf("frame output: %s", conf.FrameOutput)
}

func cycleCameraPower(pinName string) error {
	if pinName == "" {
		return nil
	}

	// It turns out when GPIO23 is low the camera's Vin->GND voltage
	// was only dropping to 2.5V instead of 0V. It seems the camera is
	// still getting some kind of power via other pins. This seems to
	// sometimes make the camera fail to reset properly (more likely
	// in some devices than others).
	//
	// Uninstalling the SPI driver disables more of the camera's pins
	// and allows the voltage to drop to 1.2V, allowing the camera to
	// reset reliably.
	//
	// Side note: uninstalling the I2C driver as well allows Vin to go
	// to 0V but we can't practically uninstall it without breaking
	// the RTC and ATtiny.
	uninstallSPIDriver()

	pin := gpioreg.ByName(pinName)

	log.Print("turning camera power off")
	if err := pin.Out(gpio.Low); err != nil {
		return fmt.Errorf("failed to set camera power pin low: %v", err)
	}
	time.Sleep(3 * time.Second)

	log.Print("turning camera power on")
	if err := pin.Out(gpio.High); err != nil {
		return fmt.Errorf("failed to set camera power pin high: %v", err)
	}

	log.Print("waiting for camera startup")
	time.Sleep(8 * time.Second)
	log.Print("camera should be ready")

	installSPIDriver()

	log.Print("host reinitialisation")
	if _, err := host.Init(); err != nil {
		return err
	}
	return nil
}

func uninstallSPIDriver() {
	log.Print("uninstalling spi driver")
	exec.Command("modprobe", "-r", "spi_bcm2835").Run()
	time.Sleep(2 * time.Second)
}

func installSPIDriver() {
	log.Print("installing spi driver")
	exec.Command("modprobe", "spi_bcm2835").Run()
	time.Sleep(8 * time.Second)
}
