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
// limitations under the License.

package cptv

import (
	"io"
	"testing"
	"time"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadV1File(t *testing.T) {
	r, err := NewFileReader("v1.cptv")
	require.NoError(t, err)
	defer r.Close()

	require.Equal(t, 1, r.Version())
	assert.Equal(t, "livingsprings03", r.DeviceName())
	assert.Equal(t, time.Date(2018, 9, 6, 9, 21, 25, 0, time.UTC), r.Timestamp().UTC().Truncate(time.Second))

	frame := new(lepton3.Frame)
	count := 0
	for {
		err := r.ReadFrame(frame)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		// Unsupported fields in v1.
		assert.Equal(t, time.Duration(0), frame.Status.TimeOn)
		assert.Equal(t, time.Duration(0), frame.Status.LastFFCTime)

		count++
	}
	assert.Equal(t, 100, count)
}
