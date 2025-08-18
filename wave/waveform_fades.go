package wave

import (
	"math"
)

func NewRollOffWaveform(wf Waveform, duration, exponent float64) Waveform {
	return &rollofWaveform{
		subwave:  wf,
		duration: duration,
		exponent: exponent,
	}
}

func NewRollInWaveform(wf Waveform, duration float64) Waveform {
	return &rollinWaveform{
		subwave:  wf,
		duration: duration,
	}
}

func NewSmoothOffWaveform(wf Waveform, duration, exponent float64) Waveform {
	return &smoothOffWaveform{
		subwave:  wf,
		duration: duration,
		exponent: exponent,
	}
}

func NewSmoothAutoOffWaveform(wf Waveform, duration, exponent float64, offAfter float64) Waveform {
	return &smoothOffWaveform{
		subwave:      wf,
		duration:     duration,
		exponent:     exponent,
		autoOffAfter: offAfter,
	}
}

func (wf *smoothOffWaveform) OffNow() {
	if wf.isOff {
		return
	}
	wf.isOff = true
	wf.subwave = NewRollOffWaveform(wf.subwave, wf.duration, wf.exponent)
}

func (wf *smoothOffWaveform) Next(deltaTime float64) (float64, bool) {
	wf.t += deltaTime
	if wf.autoOffAfter > 0 && !wf.isOff && wf.t >= wf.autoOffAfter {
		wf.OffNow()
	}
	return wf.subwave.Next(deltaTime)
}

type smoothOffWaveform struct {
	subwave      Waveform
	duration     float64
	exponent     float64
	isOff        bool
	t            float64
	autoOffAfter float64
}

type rollofWaveform struct {
	subwave  Waveform
	duration float64
	exponent float64
	t        float64
}

// OffNow implements Waveform.
func (wf *rollofWaveform) OffNow() {
	wf.subwave.OffNow()
}

func (wf *rollofWaveform) Next(deltaTime float64) (float64, bool) {
	subVal, subDone := wf.subwave.Next(deltaTime)
	wf.t += deltaTime
	var vol float64
	if wf.t >= wf.duration {
		subDone = true
		vol = 0
	} else {
		t := 1 - (wf.t / wf.duration)
		t = math.Pow(t, wf.exponent)
		vol = t
	}
	return subVal * vol, subDone
}

type rollinWaveform struct {
	subwave  Waveform
	duration float64
	t        float64
}

// OffNow implements Waveform.
func (wf *rollinWaveform) OffNow() {
	wf.subwave.OffNow()
}

func (wf *rollinWaveform) Next(deltaTime float64) (float64, bool) {
	subVal, subDone := wf.subwave.Next(deltaTime)
	wf.t += deltaTime
	var vol float64
	if wf.t >= wf.duration {
		vol = 1
	} else {
		t := wf.t / wf.duration
		vol = t
	}
	return subVal * vol, subDone
}
