package recorder

import "github.com/TheCacophonyProject/lepton3"

type Recorder interface {
	StopRecording() error
	StartRecording() error
	WriteFrame(*lepton3.Frame) error
	CheckCanRecord() error
}

type NoWriteRecorder struct {
}

func (*NoWriteRecorder) StopRecording() error            { return nil }
func (*NoWriteRecorder) StartRecording() error           { return nil }
func (*NoWriteRecorder) WriteFrame(*lepton3.Frame) error { return nil }
func (*NoWriteRecorder) CheckCanRecord() error           { return nil }
