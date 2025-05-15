package ebur128_test

import (
	"fmt"
	"math"

	"github.com/oaiiae/ebur128-go"
)

func ExampleGetVersion() {
	fmt.Println(ebur128.GetVersion())

	// Output:
	// 1 2 6
}

func fill[T int16 | int32 | float32 | float64](samples []T, amplitude, periods float64) []T {
	for i := range samples {
		theta := float64(i) / float64(len(samples)) * 2 * math.Pi * periods
		samples[i] = T(amplitude * math.Sin(theta))
	}
	return samples
}

func ExampleState_LoudnessMomentary() {
	s, err := ebur128.Init(1, 44100, ebur128.ModeM)
	if err != nil {
		panic(err)
	}
	defer s.Destroy()

	samples := fill(make([]float64, 20000), 20, 4)
	err = s.AddFramesDouble(samples, len(samples))
	if err != nil {
		panic(err)
	}

	lufs, err := s.LoudnessMomentary()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%.6f\n", lufs)

	// Output:
	// -3.535723
}
