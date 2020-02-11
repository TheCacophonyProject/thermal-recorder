package main

import (
	"encoding/binary"

	"github.com/TheCacophonyProject/go-cptv/cptvframe"
)

func convertRawBosonFrame(raw []byte, out *cptvframe.Frame) error {
	// TODO populate telemetry once bosond is sending it
	out.Status = cptvframe.Telemetry{}

	i := 0
	for y, row := range out.Pix {
		for x := range row {
			out.Pix[y][x] = binary.LittleEndian.Uint16(raw[i : i+2])
			i += 2
		}
	}
	return nil
}
