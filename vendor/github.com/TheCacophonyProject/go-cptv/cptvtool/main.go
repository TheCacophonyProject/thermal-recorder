package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/TheCacophonyProject/go-cptv"
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
	file, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}
	defer file.Close()
	bfile := bufio.NewReader(file)
	r, err := cptv.NewParser(bfile)
	if err != nil {
		return err
	}
	fields, err := r.Header()
	if err != nil {
		return err
	}

	fmt.Println(fields.Timestamp(cptv.Timestamp))
	fmt.Println(fields.Uint32(cptv.XResolution))
	fmt.Println(fields.Uint32(cptv.YResolution))
	fmt.Println(fields.Uint8(cptv.Compression))
	fmt.Println(fields.String(cptv.DeviceName))

	frames := 0
	for {
		_, _, err := r.Frame()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		frames++
	}
	fmt.Println("frames:", frames)
	return nil
}
