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

package motion

type MotionConfig struct {
	TempThresh      uint16 `yaml:"temp-thresh"`
	DeltaThresh     uint16 `yaml:"delta-thresh"`
	CountThresh     int    `yaml:"count-thresh"`
	FrameCompareGap int    `yaml:"frame-compare-gap"`
	UseOneDiffOnly  bool   `yaml:"one-diff-only"`
	TriggerFrames   int    `yaml:"trigger-frames"`
	WarmerOnly      bool   `yaml:"warmer-only"`
	Verbose         bool   `yaml:"verbose"`
}

func DefaultMotionConfig() MotionConfig {
	return MotionConfig{
		TempThresh:      2900,
		DeltaThresh:     50,
		CountThresh:     3,
		FrameCompareGap: 45,
		Verbose:         false,
		TriggerFrames:   2,
		UseOneDiffOnly:  true,
		WarmerOnly:      true,
	}
}

func (conf *MotionConfig) Validate() error {
	// TODO
	return nil
}
