package wave

import (
	"sync"
)

func NewAddWaveform(persistent bool) *AddWaveform {
	return &AddWaveform{
		waves:      nil,
		persistent: persistent,
		lock:       &sync.Mutex{},
	}
}

type AddWaveform struct {
	waves      []Waveform
	persistent bool
	lock       *sync.Mutex
}

func (wf *AddWaveform) Next(deltaTime float64) (float64, bool) {
	total := 0.0
	wf.lock.Lock()
	defer wf.lock.Unlock()
	newWaves := make([]Waveform, 0, len(wf.waves))
	for _, wf := range wf.waves {
		waveVal, waveDone := wf.Next(deltaTime)
		total += waveVal
		if !waveDone {
			newWaves = append(newWaves, wf)
		}
	}
	wf.waves = newWaves
	return total, len(wf.waves) > 0 || wf.persistent
}

func (wf *AddWaveform) Add(wf2 Waveform) {
	wf.lock.Lock()
	wf.waves = append(wf.waves, wf2)
	wf.lock.Unlock()
}

func (wf *AddWaveform) OffNow() {
	for _, w := range wf.waves {
		w.OffNow()
	}
}
