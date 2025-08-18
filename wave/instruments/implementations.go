package instruments

import (
	"encoding/json"
	"fmt"
	"wave"
)

type waveformInstrument struct {
	WaveType       SynthWaveType
	WaveRunInTime  float64
	WaveRunOutTime float64
	WaveRunOutExpo float64
	NoteOffTime    float64
	NoteOffExpo    float64
	Volume         float64
}

func (ins *waveformInstrument) CreateWaveform(key wave.Key, octave int, duration float64) wave.Waveform {
	freq := wave.NoteToFreq(key, octave)
	var wf wave.Waveform
	switch ins.WaveType {
	case SinWaveType:
		wf = wave.NewSinWaveform(freq, ins.Volume)
	case SquareWaveType:
		wf = wave.NewSquareWaveform(freq, ins.Volume, 0.25)
	case TriangleWaveType:
		wf = wave.NewTriangleWaveform(freq, ins.Volume)
	case SawtoothWaveType:
		wf = wave.NewSawtoothWaveform(freq, ins.Volume)
	default:
		panic("invalid wave type")
	}
	wf = wave.NewRollOffWaveform(wf, ins.WaveRunOutTime, ins.WaveRunOutExpo)
	wf = wave.NewRollInWaveform(wf, ins.WaveRunInTime)
	if duration == 0 {
		wf = wave.NewSmoothOffWaveform(wf, ins.NoteOffTime, ins.NoteOffExpo)
	} else {
		wf = wave.NewSmoothAutoOffWaveform(wf, ins.NoteOffTime, ins.NoteOffExpo, duration)
	}
	wf = wave.NewVolumeOscilatorWaveform(wf, wave.NewZeroSquareWaveform(15, 1, 0))
	return wf
}

type wavInstrument struct {
	WavData     *wave.WavData
	RefKey      wave.Key
	RefOctave   int
	NoteOffTime float64
	NoteOffExpo float64
	WavOffset   float64
}

func (ins *wavInstrument) CreateWaveform(key wave.Key, octave int, duration float64) wave.Waveform {
	var wf wave.Waveform
	wf = wave.NewWavWaveform(ins.WavData)
	if ins.WavOffset > 0 {
		wf.Next(ins.WavOffset)
	}
	wf = wave.NewChangeSpeedWaveform(
		wf,
		wave.SpeedChange(ins.RefKey, ins.RefOctave, key, octave),
	)
	if duration == 0 {
		wf = wave.NewSmoothOffWaveform(wf, ins.NoteOffTime, ins.NoteOffExpo)
	} else {
		wf = wave.NewSmoothAutoOffWaveform(wf, ins.NoteOffTime, ins.NoteOffExpo, duration)
	}
	return wf
}

type volumeOscilatedInstrument struct {
	ins       Instrument
	Frequency float64
	WaveType  SynthWaveType
}

func (v *volumeOscilatedInstrument) CreateWaveform(key wave.Key, octave int, duration float64) wave.Waveform {
	wf := v.ins.CreateWaveform(key, octave, duration)
	var volWf wave.Waveform
	switch v.WaveType {
	case SinWaveType:
		volWf = wave.NewSinWaveform(v.Frequency, 1)
	case SquareWaveType:
		volWf = wave.NewSquareWaveform(v.Frequency, 1, 0)
	case SawtoothWaveType:
		volWf = wave.NewSawtoothWaveform(v.Frequency, 1)
	case TriangleWaveType:
		volWf = wave.NewTriangleWaveform(v.Frequency, 1)
	default:
		panic("oof")
	}
	wf = wave.NewVolumeOscilatorWaveform(wf, volWf)
	return wf
}

type SynthWaveType uint8

const (
	NotSpecifiedWaveType SynthWaveType = iota
	SinWaveType
	SquareWaveType
	TriangleWaveType
	SawtoothWaveType
)

var synthWaveFromString = map[string]SynthWaveType{
	"sin":      SinWaveType,
	"square":   SquareWaveType,
	"triangle": TriangleWaveType,
	"sawtooth": SawtoothWaveType,
	"":         NotSpecifiedWaveType,
}

func (w *SynthWaveType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, ok := synthWaveFromString[s]
	if !ok {
		return fmt.Errorf("invalid SynthWaveType: %s", s)
	}
	*w = val
	return nil
}
