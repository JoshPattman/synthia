package wave

import "math"

func NewLowPassWaveform(wf Waveform, cutoffFreq float64) Waveform {
	return &lowPassWaveform{
		wave:       wf,
		cutoffFreq: cutoffFreq,
	}
}

func NewChangeSpeedWaveform(wf Waveform, multipler float64) Waveform {
	return &changeSpeedWaveform{
		multipler: multipler,
		wf:        wf,
	}
}

func NewVolumeOscilatorWaveform(wf Waveform, vol Waveform) Waveform {
	return &volumeOscilatorWaveform{wf: wf, volume: vol}
}

type lowPassWaveform struct {
	wave         Waveform
	currentValue float64
	cutoffFreq   float64
}

func (l *lowPassWaveform) Next(deltaTime float64) (float64, bool) {
	val, done := l.wave.Next(deltaTime)

	// compute smoothing factor (alpha) from cutoff frequency
	rc := 1.0 / (2 * math.Pi * l.cutoffFreq) // time constant
	alpha := deltaTime / (rc + deltaTime)    // standard RC filter formula

	l.currentValue += alpha * (val - l.currentValue)
	return l.currentValue, done
}

func (l lowPassWaveform) OffNow() {
	l.wave.OffNow()
}

type changeSpeedWaveform struct {
	multipler float64
	wf        Waveform
}

func (wf *changeSpeedWaveform) Next(deltaTime float64) (float64, bool) {
	return wf.wf.Next(deltaTime * wf.multipler)
}

func (wf changeSpeedWaveform) OffNow() {
	wf.wf.OffNow()
}

type volumeOscilatorWaveform struct {
	wf     Waveform
	volume Waveform
}

func (wf *volumeOscilatorWaveform) Next(deltaTime float64) (float64, bool) {
	vol, _ := wf.volume.Next(deltaTime)
	point, done := wf.wf.Next(deltaTime)
	return point * vol, done
}

func (wf volumeOscilatorWaveform) OffNow() {
	wf.wf.OffNow()
}
