package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	cptv "github.com/TheCacophonyProject/go-cptv"
	"github.com/TheCacophonyProject/lepton3"
)

func NewCPTVFileRecorder(config *Config) *CPTVFileRecorder {
	writer := new(CPTVFileRecorder)
	writer.outputDir = config.OutputDir
	writer.conf = config
	return writer
}

type CPTVFileRecorder struct {
	writer    *cptv.FileWriter
	outputDir string
	conf      *Config
}

func (cfr *CPTVFileRecorder) CheckCanRecord() error {
	enoughSpace, err := checkDiskSpace(cfr.conf.MinDiskSpace, cfr.outputDir)
	if err != nil {
		return fmt.Errorf("Problem with checking disk space: %v", err)
	} else if !enoughSpace {
		return errors.New("Motion detected but not enough free disk space to start recording")
	}
	return nil
}

func (fw *CPTVFileRecorder) StartRecording() error {
	filename := filepath.Join(fw.outputDir, newRecordingTempName())
	log.Printf("recording started: %s", filename)

	var err error
	if fw.writer, err = cptv.NewFileWriter(filename); err != nil {
		return err
	}

	if err = fw.writer.WriteHeader(fw.conf.DeviceName); err != nil {
		fw.Stop()
	}

	return err
}

func (fw *CPTVFileRecorder) StopRecording() error {
	if fw.writer != nil {
		fw.writer.Close()

		finalName, err := renameTempRecording(fw.writer.Name())
		log.Printf("recording stopped: %s\n", finalName)
		fw.writer = nil

		return err
	}
	return nil
}

func (fw *CPTVFileRecorder) Stop() {
	if fw.writer != nil {
		fw.writer.Close()
		os.Remove(fw.writer.Name())
		fw.writer = nil
	}
}

func (fw *CPTVFileRecorder) WriteFrame(frame *lepton3.Frame) error {
	return fw.writer.WriteFrame(frame)
}

func newRecordingTempName() string {
	return time.Now().Format("20060102.150405.000." + cptvTempExt)
}

func renameTempRecording(tempName string) (string, error) {
	finalName := recordingFinalName(tempName)
	err := os.Rename(tempName, finalName)
	if err != nil {
		return "", err
	}
	return finalName, nil
}

var reTempName = regexp.MustCompile(`(.+)\.temp$`)

func recordingFinalName(filename string) string {
	return reTempName.ReplaceAllString(filename, `$1`)
}

func deleteTempFiles(directory string) error {
	matches, _ := filepath.Glob(filepath.Join(directory, "*."+cptvTempExt))
	for _, filename := range matches {
		if err := os.Remove(filename); err != nil {
			return err
		}
	}
	return nil
}

func checkDiskSpace(mb uint64, dir string) (bool, error) {
	var fs syscall.Statfs_t
	if err := syscall.Statfs(dir, &fs); err != nil {
		return false, err
	}
	return fs.Bavail*uint64(fs.Bsize)/1024/1024 >= mb, nil
}
