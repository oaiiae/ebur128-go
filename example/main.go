package main

import (
	"fmt"
	"io"
	"os"
	"time"
	"unsafe"

	"github.com/oaiiae/ebur128-go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: ", os.Args[0], "FILENAME...")
		return
	}

	var r io.ReadCloser
	switch os.Args[1] {
	case "-":
		r = os.Stdin
	default:
		var err error
		r, err = os.Open(os.Args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	defer r.Close()

	const channels = 2
	const sampling = 44100

	st, err := ebur128.Init(channels, sampling, ebur128.ModeM)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer st.Destroy()

	go func() {
		for {
			time.Sleep(400 * time.Millisecond)
			fmt.Print("\r")
			fmt.Print(st.LoudnessMomentary())
		}
	}()

	src := make([]float64, channels)
	buf := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(src))), 8*len(src))
	for {
		_, err := io.ReadFull(r, buf)
		switch err {
		case io.EOF:
			return
		case nil:
		default:
			fmt.Println(err)
			return
		}
		st.AddFramesDouble(src, len(src)/channels)
	}
}
