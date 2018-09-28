// Copyright 2018 The Cacophony Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package main

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/TheCacophonyProject/go-cptv"
	"github.com/TheCacophonyProject/lepton3"
)

func main() {
	err := runMain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runMain() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: %s <filename>", os.Args[0])
	}

	fr, err := cptv.NewFileReader(os.Args[1])
	if err != nil {
		return err
	}
	defer fr.Close()

	fmt.Println("Timestamp:   ", fr.Timestamp())
	fmt.Println("Device Name: ", fr.DeviceName())

	// Read the frames and get a frame count. This is an illustration of
	// frame reading - the r.FrameCount method will do the same thing (and
	// will similarly leave the file pointer at EOF)
	frames := 0
	frame := new(lepton3.Frame)
	for {
		err := fr.ReadFrame(frame)
		if err != nil {
			if err == io.EOF {
				fmt.Print(".")
				frames++ // the last valid read returns EOF
				break
			}
			return err
		}
		frames++
		fmt.Print(".")
	}
	fmt.Print("\n")
	fmt.Println("Frame Count: ", frames)

	return nil
}
