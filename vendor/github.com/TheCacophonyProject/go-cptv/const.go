package cptv

const (
	magic        = "CPTV"
	version byte = 0x01

	headerSection = 'H'
	frameSection  = 'F'

	// Header field keys
	Timestamp   byte = 'T'
	XResolution byte = 'X'
	YResolution byte = 'Y'
	Compression byte = 'C'
	DeviceName  byte = 'D'

	// Frame field keys
	Offset    byte = 't'
	BitWidth  byte = 'w'
	FrameSize byte = 'f'
)
