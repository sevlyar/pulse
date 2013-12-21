package main

import (
	"fmt"
	"github.com/mjibson/go-dsp/fft"
	"log"
	"math/cmplx"
	"os"
	"os/signal"
	"strings"
)

const STRLEN = 63

var (
	PEACKBAR = strings.Repeat("=", STRLEN-1) + "@"
	SPACEBAR = strings.Repeat(" ", STRLEN)
)

var hist = &Hist{
	SampleRate: 44100,
	BufferLen:  1024,
	Bars: []*Bar{
		&Bar{Color: "31", Set: []int{0}},
		&Bar{Color: "31", Set: []int{1, 2}},
		&Bar{Color: "33", Set: []int{4, -10}},
		&Bar{Color: "33", Set: []int{11, -17}},
		&Bar{Color: "32", Set: []int{18, -45}},
		&Bar{Color: "32", Set: []int{46, -80}},
		&Bar{Color: "34", Set: []int{81, -290}},
		&Bar{Color: "34", Set: []int{291, -510}},
	},
	FilterA: 0.01,
	FilterN: 2,
}

func main() {
	hist.Init()

	ss := NewPaSampleSpec(PA_SAMPLE_S16LE, hist.SampleRate, 1)
	simple, err := NewPaSimple("", "dspvis", PA_STREAM_RECORD, "", "record", ss)
	if err != nil {
		log.Panicln("pa_simple_new error:", err)
	}
	defer simple.Free()

	fmt.Println("Recording.  Press Ctrl-C to stop.")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	buf := make([]int16, hist.BufferLen)
READ_CYCLE:
	for {
		if _, err = simple.ReadBuffer(buf); err != nil {
			log.Panicln("pa_simple_read error:", err)
		}
		process(buf)

		select {
		case <-sig:
			break READ_CYCLE
		default:
		}
	}

	fmt.Println()
}

func process(in []int16) {
	samples := spectre(in)
	hist.Update(samples)
	hist.Draw(drawBar)
	fmt.Printf("\x1B[%dA", len(hist.Bars))
}

func spectre(in []int16) []float64 {
	samples := make([]float64, len(in))
	for i, inval := range in {
		samples[i] = float64(inval)
	}

	fftcomp := fft.FFTReal(samples)

	for i, comp := range fftcomp {
		samples[i] = cmplx.Abs(comp)
	}

	return samples
}

func drawBar(bar *Bar) {
	lvl, avg := bar.Levels()

	l := int(STRLEN * lvl)
	if l < 0 {
		l = 0
	}
	fmt.Printf("\x1B[%sm%s%s", bar.Color, PEACKBAR[:l], SPACEBAR[l:])

	l = int(STRLEN * avg)
	fmt.Printf("\r\x1B[%dC|\n\x1B[0m", l)
}
