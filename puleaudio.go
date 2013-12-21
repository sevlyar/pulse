package main

/*
#cgo LDFLAGS: -lpulse-simple -lpulse
#include <pulse/simple.h>
#include <pulse/error.h>
#include <pulse/sample.h>
*/
import "C"
import (
	"errors"
	"reflect"
	"unsafe"
)

const (
	PA_SAMPLE_U8 = iota
	PA_SAMPLE_ALAW
	PA_SAMPLE_ULAW
	PA_SAMPLE_S16LE
	PA_SAMPLE_S16BE
	PA_SAMPLE_FLOAT32LE
	PA_SAMPLE_FLOAT32BE
	PA_SAMPLE_S32LE
	PA_SAMPLE_S32BE
	PA_SAMPLE_S24LE
	PA_SAMPLE_S24BE
	PA_SAMPLE_S24_32LE
	PA_SAMPLE_S24_32BE
	PA_SAMPLE_MAX
	PA_SAMPLE_INVALID = -1
)

const (
	PA_STREAM_NODIRECTION = iota
	PA_STREAM_PLAYBACK
	PA_STREAM_RECORD
	PA_STREAM_UPLOAD
)

type PaSampleSpec *C.pa_sample_spec

func NewPaSampleSpec(format, rate, channels int) (ss PaSampleSpec) {
	return PaSampleSpec(&C.pa_sample_spec{
		format:   C.pa_sample_format_t(format),
		rate:     C.uint32_t(rate),
		channels: C.uint8_t(channels)})
}

type PaSimple struct {
	s *C.pa_simple
}

func NewPaSimple(server, name string, direction int, device, streamName string, sampleSpec PaSampleSpec) (simple *PaSimple, err error) {
	var e C.int
	s := C.pa_simple_new(
		strGo2C(server),
		strGo2C(name),
		C.pa_stream_direction_t(direction),
		strGo2C(device),
		strGo2C(streamName),
		sampleSpec,
		nil,
		nil,
		&e)
	if s == nil {
		err = PaError(e)
	} else {
		simple = &PaSimple{s}
	}
	return
}

func strGo2C(gostr string) *C.char {
	if len(gostr) > 0 {
		return C.CString(gostr)
	}
	return nil
}

func (simple *PaSimple) Free() {
	C.pa_simple_free(simple.s)
}

func (simple *PaSimple) Read(p []byte) (n int, err error) {
	var e C.int
	n = int(C.pa_simple_read(simple.s, unsafe.Pointer(&p[0]), C.size_t(len(p)), &e))
	if n < 0 {
		err = PaError(e)
	}
	return
}

func (simple *PaSimple) ReadBuffer(p interface{}) (n int, err error) {
	val := reflect.ValueOf(p)

	var e C.int
	n = int(C.pa_simple_read(simple.s, unsafe.Pointer(val.Pointer()), C.size_t(val.Cap()), &e))
	if n < 0 {
		err = PaError(e)
	}
	return
}

func PaError(e C.int) (err error) {
	cstr := C.pa_strerror(e)
	gostr := C.GoString(cstr)
	return errors.New(gostr)
}
