# go-cptv

[![Build Status](https://api.travis-ci.com/TheCacophonyProject/go-cptv.svg?branch=master)](https://travis-ci.com/TheCacophonyProject/go-cptv)

This package implements a Go package for generating and consuming
Cacophony Project Thermal Video (CPTV) files. For more details on
these files see the [specification](https://github.com/TheCacophonyProject/go-cptv/blob/master/SPEC.md).


## Example Usage

### Writing CPTV Files

```go

import (
    "github.com/TheCacophonyProject/go-cptv"
    "github.com/TheCacophonyProject/lepton3"
)


func writeFrames(frames []*lepton3.Frame) error {
    w := cptv.NewFileWriter("out.cptv")
    defer w.Close()
    err := w.WriterHeader("device-name")
    if err != nil {
        return err
    }
    for _, frame := range frames {
        err := w.WriteFrame(frame)
        if err != nil {
            return err
        }
    }
    return nil
}
```

### Reading CPTV Files

```go

import (
    "fmt"
    "os"

    "github.com/TheCacophonyProject/go-cptv"
    "github.com/TheCacophonyProject/lepton3"
)


func readFrames() ([]*lepton3.Frame, error) {
    f, err := os.Open("some.cptv")
    if err != nil {
        return nil, err
    }
    defer f.Close()

    r, err := cptv.NewReader(f)
    if err != nil {
        return nil, err
    }
    fmt.Println("timestamp:", r.Timestamp())
    fmt.Println("device:", r.DeviceName())

    var out []*lepton3.Frame
    for {
        frame := new(lepton3.Frame)
        err := r.ReadFrame(frame)
        if err != nil {
            if err == io.EOF {
                return out, nil
            }
            return err
        }
        out = append(out, frame)
    }
}
```
