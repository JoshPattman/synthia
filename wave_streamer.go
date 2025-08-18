package main

import (
	"wave"

	"github.com/gopxl/beep/v2"
)

type WaveStreamer struct {
	wf wave.Waveform
}

func (ws *WaveStreamer) SetWave(wf wave.Waveform) {
	ws.wf = wf
}

func (ws *WaveStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	timePerSample := beep.SampleRate(44100).D(1).Seconds()
	for i := range samples {
		value, _ := ws.wf.Next(timePerSample)
		samples[i][0] = value
		samples[i][1] = value
	}
	return len(samples), true
}

func (*WaveStreamer) Err() error {
	return nil
}
