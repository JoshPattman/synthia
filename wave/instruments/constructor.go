package instruments

import (
	"errors"
	"fmt"
	"wave"
)

type ConstructorJson struct {
	Source struct {
		Type           string        `json:"type"`              // Either "generate" or "wav"
		WaveType       SynthWaveType `json:"wave_type"`         // One of square, sawtooth, triangle, sin
		WaveRunInTime  float64       `json:"wave_run_in_time"`  // For waves only
		WaveRunOutTime float64       `json:"wave_run_out_time"` // For waves only
		WaveRunOutExpo float64       `json:"wave_run_out_expo"` // For waves only
		WavPath        string        `json:"wav_path"`          // Path to wav file
		WavOffset      float64       `json:"wav_offset"`        // How many seconds should we offset the wav file by?
		WavRefNote     wave.Key      `json:"wav_note"`          // The default wav key
		WavRefOctave   int           `json:"wav_octave"`        // The default wav octave
	} `json:"source"`
	Volume          float64 `json:"volume"` // Volume multipler
	NoteOffTime     float64 `json:"note_off_time"`
	NoteOffExpo     float64 `json:"note_off_expo"`
	VolumeOscilator struct {
		Frequency float64       `json:"frequency"`
		WaveType  SynthWaveType `json:"wave_type"`
	} `json:"volume_oscilator"`
}

type Instrument interface {
	CreateWaveform(key wave.Key, octave int, duration float64) wave.Waveform
}

func NewInstrument(cons ConstructorJson) (Instrument, error) {
	var ins Instrument
	switch cons.Source.Type {
	case "generate":
		ins = &waveformInstrument{
			WaveType:       cons.Source.WaveType,
			WaveRunInTime:  cons.Source.WaveRunInTime,
			WaveRunOutTime: cons.Source.WaveRunOutTime,
			WaveRunOutExpo: cons.Source.WaveRunOutExpo,
			NoteOffTime:    cons.NoteOffTime,
			NoteOffExpo:    cons.NoteOffExpo,
			Volume:         cons.Volume,
		}
	case "wav":
		wavData, err := wave.NewWavDataFromFile(cons.Source.WavPath)
		if err != nil {
			return nil, errors.Join(errors.New("failed to read wav file"), err)
		}
		ins = &wavInstrument{
			WavData:     wavData,
			RefKey:      cons.Source.WavRefNote,
			RefOctave:   cons.Source.WavRefOctave,
			NoteOffTime: cons.NoteOffTime,
			NoteOffExpo: cons.NoteOffExpo,
			WavOffset:   cons.Source.WavOffset,
		}
	default:
		return nil, fmt.Errorf("unrecognised source type '%s'", cons.Source.Type)
	}
	if cons.VolumeOscilator.Frequency != 0 {
		ins = &volumeOscilatedInstrument{ins, cons.VolumeOscilator.Frequency, cons.VolumeOscilator.WaveType}
	}
	return ins, nil
}
