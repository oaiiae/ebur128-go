// Simple Go translation of [libebur128]'s minimal-example.
//
// [libebur128]: https://github.com/jiixyj/libebur128
package main

/*
#cgo LDFLAGS: -lsndfile
#include <stdlib.h>
#include <sndfile.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"

	"github.com/oaiiae/ebur128-go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: ", os.Args[0], "FILENAME...")
		return
	}

	sts := make(ebur128.States, len(os.Args)-1)
	for i, arg := range os.Args[1:] {
		filepath := C.CString(arg)
		defer C.free(unsafe.Pointer(filepath))

		var file_info C.SF_INFO
		file := C.sf_open(filepath, C.SFM_READ, &file_info)
		if file == nil {
			fmt.Println("Could not create ebur128_state!")
			return
		}
		defer C.sf_close(file)

		st, err := ebur128.Init(uint(file_info.channels),
			uint64(file_info.samplerate), ebur128.ModeI)
		if err != nil {
			fmt.Println("Could not create ebur128_state!", "err", err)
			return
		}
		defer st.Destroy()
		sts[i] = st

		// example: set channel map (note: see ebur128.h for the default map)
		if file_info.channels == 5 {
			st.SetChannel(0, ebur128.Left)
			st.SetChannel(1, ebur128.Right)
			st.SetChannel(2, ebur128.Center)
			st.SetChannel(3, ebur128.LeftSurround)
			st.SetChannel(4, ebur128.RightSurround)
		}

		buffer := make([]float64, file_info.samplerate*file_info.channels)

		for {
			nr_frames_read := C.sf_readf_double(
				file, (*C.double)(unsafe.SliceData(buffer)), C.sf_count_t(file_info.samplerate))
			if nr_frames_read == 0 {
				break
			}
			st.AddFramesDouble(buffer, int(nr_frames_read))
		}

		loudness, err := st.LoudnessGlobal()
		fmt.Printf("%.2f LUFS, %s (%v)\n", loudness, arg, err)
	}

	loudness, err := sts.LoudnessGlobal()
	fmt.Printf("-----------\n%.2f LUFS (%v)\n", loudness, err)
}
