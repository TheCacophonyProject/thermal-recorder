package main

import (
	"log"
	"net"
	"os"

	"github.com/TheCacophonyProject/lepton3"
	arg "github.com/alexflint/go-arg"
	"periph.io/x/periph/host"
)

const (
	framesHz    = 9 // approx
	cptvTempExt = "cptv.temp"

	frameLogIntervalFirstMin = 15 * framesHz
	frameLogInterval         = 60 * 5 * framesHz
)

var (
	version   = "<not set>"
	frameLoop *FrameLoop
)

type Args struct {
	ConfigFile         string `arg:"-c,--config" help:"path to configuration file"`
	UploaderConfigFile string `arg:"-u,--uploader-config" help:"path to uploader config file"`
	Timestamps         bool   `arg:"-t,--timestamps" help:"include timestamps in log output"`
	TestCptvFile       string `arg:"-f, --testfile" help:"Run a CPTV file through to see what the results are"`
	Verbose            bool   `arg:"-v, --verbose" help:"Make logging more verbose"`
}

func (Args) Version() string {
	return version
}

func procArgs() Args {
	var args Args
	args.ConfigFile = "/etc/thermal-recorder.yaml"
	args.UploaderConfigFile = "/etc/thermal-uploader.yaml"
	arg.MustParse(&args)
	return args
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

	log.Printf("running version: %s", version)
	conf, err := ParseConfigFiles(args.ConfigFile, args.UploaderConfigFile)
	if err != nil {
		return err
	}

	if args.TestCptvFile != "" {
		conf.Motion.Verbose = args.Verbose
		results := NewCPTVPlaybackTester(conf).Detect(args.TestCptvFile)
		log.Printf("Detected: %-16s Recorded: %-16s Motion frames: %d/%d", results.motionDetectedFrames, results.recordedFrames, results.motionDetectedCount, results.frameCount)
		return nil
	}

	logConfig(conf)

	log.Println("starting d-bus service")
	err = startService(conf.OutputDir)
	if err != nil {
		return err
	}

	log.Println("host initialisation")
	if _, err := host.Init(); err != nil {
		return err
	}

	turret := NewTurretController(conf.Turret)
	go turret.Start()

	log.Println("deleting temp files")
	if err := deleteTempFiles(conf.OutputDir); err != nil {
		return err
	}

	for {
		// Set up listener for frames sent by leptond.
		os.Remove(conf.FrameInput)
		listener, err := net.Listen("unixpacket", conf.FrameInput)
		if err != nil {
			return err
		}
		log.Print("waiting for camera connection")

		conn, err := listener.Accept()
		if err != nil {
			log.Printf("socket accept failed: %v", err)
			continue
		}

		// Prevent concurrent connections.
		listener.Close()

		err = handleConn(conn, conf, turret)
		log.Printf("camera connection ended with: %v", err)
	}
}

func handleConn(conn net.Conn, conf *Config, turret *TurretController) error {

	totalFrames := 0

	cptvRecorder := NewCPTVFileRecorder(conf)
	defer cptvRecorder.Stop()

	processor := NewMotionProcessor(conf, nil, cptvRecorder)
	frameLoop = processor.frameLoop

	rawFrame := new(lepton3.RawFrame)

	log.Print("new camera connection, reading frames")

	for {
		_, err := conn.Read(rawFrame[:])
		if err != nil {
			return err
		}
		totalFrames++

		if totalFrames%frameLogIntervalFirstMin == 0 &&
			totalFrames <= 60*framesHz || totalFrames%frameLogInterval == 0 {
			log.Printf("%d frames for this connection", totalFrames)
		}

		processor.Process(rawFrame)
	}
}

func logConfig(conf *Config) {
	log.Printf("device name: %s", conf.DeviceName)
	log.Printf("frame input: %s", conf.FrameInput)
	log.Printf("output dir: %s", conf.OutputDir)
	log.Printf("recording limits: %ds to %ds", conf.MinSecs, conf.MaxSecs)
	log.Printf("preview seconds: %d", conf.PreviewSecs)
	log.Printf("minimum disk space: %d", conf.MinDiskSpace)
	log.Printf("motion: %+v", conf.Motion)
	if !conf.WindowStart.IsZero() {
		log.Printf("recording window: %02d:%02d to %02d:%02d",
			conf.WindowStart.Hour(), conf.WindowStart.Minute(),
			conf.WindowEnd.Hour(), conf.WindowEnd.Minute())
	}
	if conf.Turret.Active {
		log.Printf("Turret active")
		log.Printf("\tPID: %v", conf.Turret.PID)
		log.Printf("\tServoX: %+v", conf.Turret.ServoX)
		log.Printf("\tServoY: %+v", conf.Turret.ServoY)
	}
}
