package main

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/TheCacophonyProject/go-cptv/cptvframe"
	"github.com/TheCacophonyProject/lepton3"
)

func convertRawBosonFrame(raw []byte, out *cptvframe.Frame, edgePixels int) error {
	// TODO populate telemetry once bosond is sending it
	out.Status = cptvframe.Telemetry{
		// Make it appear like there hasn't been a FFC recently. Without
		// this the motion detector will never trigger.
		LastFFCTime: time.Second,
		TimeOn:      time.Minute,
	}

	i := 0
	for y, row := range out.Pix {
		for x := range row {
			out.Pix[y][x] = binary.LittleEndian.Uint16(raw[i : i+2])
			onEdge := y < edgePixels || x < edgePixels || y >= (len(out.Pix)-edgePixels) || x >= (len(row)-edgePixels)
			if !onEdge && out.Pix[y][x] == 0 {
				err := fmt.Errorf("Bad pixel (%d,%d) of %d", y, x, out.Pix[y][x])
				return &lepton3.BadFrameErr{Cause: err}
			}
			i += 2
		}
	}
	return nil
}
