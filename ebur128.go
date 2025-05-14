// Package ebur128 provides Go bindings for [libebur128], a library
// for loudness measurement according to the EBU R128 standard.
//
// [libebur128]: https://github.com/jiixyj/libebur128
package ebur128

/*
#cgo LDFLAGS: -lebur128
#include <ebur128.h>
*/
import "C"
import (
	"time"
	"unsafe"
)

// Channels
//   - Use these values when setting the channel map with [State.SetChannel].
//   - See definitions in ITU R-REC-BS 1770-4
const (
	Unused        = C.EBUR128_UNUSED         // unused channel (for example LFE channel)
	Left          = C.EBUR128_LEFT           //
	Mp030         = C.EBUR128_Mp030          // itu M+030
	Right         = C.EBUR128_RIGHT          //
	Mm030         = C.EBUR128_Mm030          // itu M-030
	Center        = C.EBUR128_CENTER         //
	Mp000         = C.EBUR128_Mp000          // itu M+000
	LeftSurround  = C.EBUR128_LEFT_SURROUND  //
	Mp110         = C.EBUR128_Mp110          // itu M+110
	RightSurround = C.EBUR128_RIGHT_SURROUND //
	Mm110         = C.EBUR128_Mm110          // itu M-110
	DualMono      = C.EBUR128_DUAL_MONO      // a channel that is counted twice
	MpSC          = C.EBUR128_MpSC           // itu M+SC
	MmSC          = C.EBUR128_MmSC           // itu M-SC
	Mp060         = C.EBUR128_Mp060          // itu M+060
	Mm060         = C.EBUR128_Mm060          // itu M-060
	Mp090         = C.EBUR128_Mp090          // itu M+090
	Mm090         = C.EBUR128_Mm090          // itu M-090
	Mp135         = C.EBUR128_Mp135          // itu M+135
	Mm135         = C.EBUR128_Mm135          // itu M-135
	Mp180         = C.EBUR128_Mp180          // itu M+180
	Up000         = C.EBUR128_Up000          // itu U+000
	Up030         = C.EBUR128_Up030          // itu U+030
	Um030         = C.EBUR128_Um030          // itu U-030
	Up045         = C.EBUR128_Up045          // itu U+045
	Um045         = C.EBUR128_Um045          // itu U-030
	Up090         = C.EBUR128_Up090          // itu U+090
	Um090         = C.EBUR128_Um090          // itu U-090
	Up110         = C.EBUR128_Up110          // itu U+110
	Um110         = C.EBUR128_Um110          // itu U-110
	Up135         = C.EBUR128_Up135          // itu U+135
	Um135         = C.EBUR128_Um135          // itu U-135
	Up180         = C.EBUR128_Up180          // itu U+180
	Tp000         = C.EBUR128_Tp000          // itu T+000
	Bp000         = C.EBUR128_Bp000          // itu B+000
	Bp045         = C.EBUR128_Bp045          // itu B+045
	Bm045         = C.EBUR128_Bm045          // itu B-045
)

type ebur128Error C.int

var (
	ErrNomem               error = ebur128Error(C.EBUR128_ERROR_NOMEM)
	ErrInvalidMode         error = ebur128Error(C.EBUR128_ERROR_INVALID_MODE)
	ErrInvalidChannelIndex error = ebur128Error(C.EBUR128_ERROR_INVALID_CHANNEL_INDEX)
	ErrNoChange            error = ebur128Error(C.EBUR128_ERROR_NO_CHANGE)
)

func (e ebur128Error) Error() string {
	switch e {
	case C.EBUR128_ERROR_NOMEM:
		return "ebur128: nomem"
	case C.EBUR128_ERROR_INVALID_MODE:
		return "ebur128: invalid mode"
	case C.EBUR128_ERROR_INVALID_CHANNEL_INDEX:
		return "ebur128: invalid channel index"
	case C.EBUR128_ERROR_NO_CHANGE:
		return "ebur128: no change"
	case C.EBUR128_SUCCESS:
		return "ebur128: success"
	default:
		return "ebur128: unknown error"
	}
}

// newError returns a new [ebur128Error] from libebur128 error codes.
//
//go:inline
func newError(rc C.int) error {
	if rc == C.EBUR128_SUCCESS {
		return nil
	}
	return ebur128Error(rc)
}

// Modes
//   - Use these values in [Init] (or'ed). Try to use the lowest possible.
//   - modes that suit your needs, as performance will be better.
const (
	ModeM          = C.EBUR128_MODE_M           // can call [State.LoudnessMomentary]
	ModeS          = C.EBUR128_MODE_S           // can call [State.LoudnessShortterm]
	ModeI          = C.EBUR128_MODE_I           // can call [State.LoudnessGlobal] and [State.RelativeThreshold]
	ModeLRA        = C.EBUR128_MODE_LRA         // can call [State.LoudnessRange]
	ModeSamplePeak = C.EBUR128_MODE_SAMPLE_PEAK // can call [State.SamplePeak]
	ModeTruePeak   = C.EBUR128_MODE_TRUE_PEAK   // can call [State.TruePeak]
	ModeHistogram  = C.EBUR128_MODE_HISTOGRAM   // use histogram algorithm to calculate loudness
)

// GetVersion returns library version number.
func GetVersion() (major, minor, patch int) { //nolint: nonamedreturns // names help understanding the returned info
	var x, y, z C.int
	C.ebur128_get_version(&x, &y, &z)
	return int(x), int(y), int(z)
}

// State contains information about the state of a loudness measurement.
type State C.ebur128_state

// c is a helper method to return the underlying [C.ebur128_state].
func (s *State) c() *C.ebur128_state { return (*C.ebur128_state)(s) }

// Init returns an initialized library state.
//   - channels the number of channels.
//   - samplerate the sample rate.
//   - mode see the mode enum for possible values.
func Init(channels uint, sampleRate uint64, mode int) (*State, error) {
	s := C.ebur128_init(C.uint(channels), C.ulong(sampleRate), C.int(mode))
	if s == nil {
		return nil, ErrNomem
	}
	return (*State)(s), nil
}

// Destroy library state.
func (s *State) Destroy() {
	cst := s.c()
	C.ebur128_destroy(&cst) //nolint: gocritic // false positive, see: https://github.com/go-critic/go-critic/issues/897
}

// SetChannel sets channel type. The default is:
//   - 0 -> [Left]
//   - 1 -> [Right]
//   - 2 -> [Center]
//   - 3 -> [Unused]
//   - 4 -> [LeftSurround]
//   - 5 -> [RightSurround]
//
// Params:
//   - channelNumber zero based channel index.
//   - value the channel type.
//
// Returns [ErrInvalidChannelIndex] if invalid channel index.
func (s *State) SetChannel(channelNumber uint, value int) error {
	rc := C.ebur128_set_channel(s.c(), C.uint(channelNumber), C.int(value))
	return newError(rc)
}

// ChangeParameters changes library parameters.
//
// Note that the channel map will be reset when setting a different number of
// channels. The current unfinished block will be lost.
//
// Params:
//   - channels new number of channels.
//   - sampleRate new sample rate.
//
// Returns [ErrNomem] on memory allocation error. The state will be invalid and
// must be destroyed. [ErrNoChange] if channels and sample rate were not changed.
func (s *State) ChangeParameters(channels uint, sampleRate uint64) error {
	rc := C.ebur128_change_parameters(s.c(), C.uint(channels), C.ulong(sampleRate))
	return newError(rc)
}

// SetMaxWindow sets the maximum window duration (ms precision) that will be used for [State.LoudnessWindow].
// Note that this destroys the current content of the audio buffer.
//
// Returns [ErrNomem] on memory allocation error. The state will be invalid
// and must be destroyed. [ErrNoChange] if window duration not changed.
func (s *State) SetMaxWindow(window time.Duration) error {
	if window < 0 {
		panic("negative window duration")
	}

	rc := C.ebur128_set_max_window(s.c(), C.ulong(window.Milliseconds()))
	return newError(rc)
}

// SetMaxHistory sets the maximum history duration (ms precision) that will be stored for loudness integration.
// More history provides more accurate results, but requires more resources.
//
// Applies to [State.LoudnessRange] and [State.LoudnessGlobal] when
// [ModeHistogram] is not set.
//
// Default is ULONG_MAX (at least ~50 days).
// Minimum is 3000ms for [ModeLRA] and 400ms for [ModeM].
//
// Returns [ErrNoChange] if history not changed.
func (s *State) SetMaxHistory(history time.Duration) error {
	if history < 0 {
		panic("negative history duration")
	}

	rc := C.ebur128_set_max_history(s.c(), C.ulong(history.Milliseconds()))
	return newError(rc)
}

// AddFramesShort adds frames to be processed.
//   - src array of source frames. Channels must be interleaved.
//   - frames number of frames. Not number of samples!
//
// Returns [ErrNomem] on memory allocation error.
func (s *State) AddFramesShort(src []int16, frames uint64) error {
	rc := C.ebur128_add_frames_short(s.c(), (*C.short)(unsafe.SliceData(src)), C.ulong(frames))
	return newError(rc)
}

// See [State.AddFramesShort].
func (s *State) AddFramesInt(src []int32, frames uint64) error {
	rc := C.ebur128_add_frames_int(s.c(), (*C.int)(unsafe.SliceData(src)), C.ulong(frames))
	return newError(rc)
}

// See [State.AddFramesShort].
func (s *State) AddFramesFloat(src []float32, frames uint64) error {
	rc := C.ebur128_add_frames_float(s.c(), (*C.float)(unsafe.SliceData(src)), C.ulong(frames))
	return newError(rc)
}

// See [State.AddFramesShort].
func (s *State) AddFramesDouble(src []float64, frames uint64) error {
	rc := C.ebur128_add_frames_double(s.c(), (*C.double)(unsafe.SliceData(src)), C.ulong(frames))
	return newError(rc)
}

// LoudnessGlobal returns global integrated loudness in LUFS.
//
// Returns [ErrInvalidMode] if mode [ModeI] has not been set.
func (s *State) LoudnessGlobal() (float64, error) {
	var out C.double
	rc := C.ebur128_loudness_global(s.c(), &out)
	return float64(out), newError(rc)
}

// LoudnessMomentary returns momentary loudness (last 400ms) in LUFS.
//
// Returns [ErrInvalidMode] if mode [ModeM] has not been set.
func (s *State) LoudnessMomentary() (float64, error) {
	var out C.double
	rc := C.ebur128_loudness_momentary(s.c(), &out)
	return float64(out), newError(rc)
}

// LoudnessShortterm returns short-term loudness (last 3s) in LUFS.
//
// Returns [ErrInvalidMode] if mode [ModeS] has not been set.
func (s *State) LoudnessShortterm() (float64, error) {
	var out C.double
	rc := C.ebur128_loudness_shortterm(s.c(), &out)
	return float64(out), newError(rc)
}

// LoudnessWindow returns loudness of the specified window (ms precision) in LUFS.
//
// window must not be larger than the current window set in state.
// The current window can be changed by calling [State.SetMaxWindow].
//
// Returns [ErrInvalidMode] if window larger than current window in state.
func (s *State) LoudnessWindow(window time.Duration) (float64, error) {
	if window < 0 {
		panic("negative window duration")
	}

	var out C.double
	rc := C.ebur128_loudness_window(s.c(), C.ulong(window.Milliseconds()), &out)
	return float64(out), newError(rc)
}

// LoudnessRange returns loudness range (LRA) of program in LU.
// Calculates loudness range according to EBU 3342.
//
// Returns [ErrInvalidMode] if mode [ModeLRA] has not been set.
// [ErrNomem] on memory allocation error.
func (s *State) LoudnessRange() (float64, error) {
	var out C.double
	rc := C.ebur128_loudness_range(s.c(), &out)
	return float64(out), newError(rc)
}

// SamplePeak returns maximum sample linear peak from all frames that
// have been processed for given channelNumber.
//
// The equation to convert to dBFS is: 20 * log10(out)
//
// Returns [ErrInvalidMode] if [ModeSamplePeak] has not been set.
// [ErrInvalidChannelIndex] if invalid channel index.
func (s *State) SamplePeak(channelNumber uint) (float64, error) {
	var out C.double
	rc := C.ebur128_sample_peak(s.c(), C.uint(channelNumber), &out)
	return float64(out), newError(rc)
}

// PrevSamplePeak returns maximum sample linear peak from the last call
// to [State.AddFramesShort] (and others) for given channelNumber.
//
// The equation to convert to dBFS is: 20 * log10(out)
//
// Returns [ErrInvalidMode] if [ModeSamplePeak] has not been set.
// [ErrInvalidChannelIndex] if invalid channel index.
func (s *State) PrevSamplePeak(channelNumber uint) (float64, error) {
	var out C.double
	rc := C.ebur128_prev_sample_peak(s.c(), C.uint(channelNumber), &out)
	return float64(out), newError(rc)
}

// TruePeak returns maximum true peak from all frames that have been processed for given channelNumber.
//
// Uses an implementation defined algorithm to calculate the true peak. Do not
// try to compare resulting values across different versions of the library,
// as the algorithm may change.
//
// The current implementation uses a custom polyphase FIR interpolator to
// calculate true peak. Will oversample 4x for sample rates < 96000 Hz, 2x for
// sample rates < 192000 Hz and leave the signal unchanged for 192000 Hz.
//
// The equation to convert to dBTP is: 20 * log10(out)
//
// Returns [ErrInvalidMode] if [ModeTruePeak] has not been set.
// [ErrInvalidChannelIndex] if invalid channel index.
func (s *State) TruePeak(channelNumber uint) (float64, error) {
	var out C.double
	rc := C.ebur128_true_peak(s.c(), C.uint(channelNumber), &out)
	return float64(out), newError(rc)
}

// PrevTruePeak returns maximum true peak from the last call
// to [State.AddFramesShort] (and others) for given channelNumber.
//
// Uses an implementation defined algorithm to calculate the true peak. Do not
// try to compare resulting values across different versions of the library,
// as the algorithm may change.
//
// The current implementation uses a custom polyphase FIR interpolator to
// calculate true peak. Will oversample 4x for sample rates < 96000 Hz, 2x for
// sample rates < 192000 Hz and leave the signal unchanged for 192000 Hz.
//
// The equation to convert to dBTP is: 20 * log10(out)
//
// Returns [ErrInvalidMode] if [ModeTruePeak] has not been set.
// [ErrInvalidChannelIndex] if invalid channel index.
func (s *State) PrevTruePeak(channelNumber uint) (float64, error) {
	var out C.double
	rc := C.ebur128_prev_true_peak(s.c(), C.uint(channelNumber), &out)
	return float64(out), newError(rc)
}

// RelativeThreshold returns relative threshold in LUFS.
//
// Returns [ErrInvalidMode] if mode [ModeI] has not been set.
func (s *State) RelativeThreshold() (float64, error) {
	var out C.double
	rc := C.ebur128_relative_threshold(s.c(), &out)
	return float64(out), newError(rc)
}

type States []*State

// c is a helper method to return the [States] as a "slice" of [C.ebur128_state].
func (sts States) c() (**C.ebur128_state, C.size_t) {
	ptr, size := unsafe.SliceData(sts), len(sts)
	return (**C.ebur128_state)(unsafe.Pointer(ptr)), C.size_t(size)
}

// LoudnessGlobal returns global integrated loudness in LUFS across multiple instances.
//
// Returns [ErrInvalidMode] if mode [ModeI] has not been set.
func (sts States) LoudnessGlobal() (float64, error) {
	states, size := sts.c()
	var out C.double
	rc := C.ebur128_loudness_global_multiple(states, size, &out)
	return float64(out), newError(rc)
}

// LoudnessRange returns loudness range (LRA) in LU across multiple instances.
// Calculates loudness range according to EBU 3342.
//
// Returns [ErrInvalidMode] if mode [ModeLRA] has not been set.
// [ErrNomem] on memory allocation error.
func (sts States) LoudnessRange() (float64, error) {
	states, size := sts.c()
	var out C.double
	rc := C.ebur128_loudness_range_multiple(states, size, &out)
	return float64(out), newError(rc)
}
