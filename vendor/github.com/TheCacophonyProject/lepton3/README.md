# lepton3

This package implements a robust, efficient Go interface to the FLIR
Lepton 3 thermal cameras. Care is taken to avoid memory allocations in
the critical path where possible to maximise performance.

## Usage

A reasonably simple example of how to use the package can be found in
the `leptonutil` subdirectory. This is a utility which captures
frames from the camera and optionally saves them to PNG files.

To build leptonutil for the Raspberry Pi:

```
$ export GOARCH=arm
$ export GOARM=7
$ cd leptonutil
$ go build    # not install!
```

## Further information

The data sheet for the camera contains a wealth of useful
information. It can be found here:
http://www.flir.com/uploadedFiles/OEM/Products/LWIR-Cameras/Lepton/Lepton-3-Engineering-Datasheet.pdf
