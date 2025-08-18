package wave

import (
	"io"
	"os"

	"github.com/youpy/go-riff"
	"github.com/youpy/go-wav"
)

type WavData struct {
	points   []float64
	duration float64
}

func NewWavDataFromFile(file string) (*WavData, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return NewWavData(f)
}

func NewWavData(r riff.RIFFReader) (*WavData, error) {
	wavReader := wav.NewReader(r)
	allSamples := []float64{}
	for {
		samples, err := wavReader.ReadSamples()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		for _, s := range samples {
			allSamples = append(allSamples, wavReader.FloatValue(s, 0))
		}
	}
	dur, err := wavReader.Duration()
	if err != nil {
		return nil, err
	}
	return &WavData{
		points:   allSamples,
		duration: dur.Seconds(),
	}, nil
}

func NewWavWaveform(d *WavData) Waveform {
	return &wavWaveform{data: d}
}

type wavWaveform struct {
	t       float64
	data    *WavData
	stopped bool
}

func (wf *wavWaveform) OffNow() {
	wf.stopped = true
}

func (wf *wavWaveform) Next(deltaTime float64) (float64, bool) {
	wf.t += deltaTime
	if wf.t >= wf.data.duration || wf.stopped {
		return 0, true
	}
	index := int(wf.t / wf.data.duration * float64(len(wf.data.points)))
	return wf.data.points[index], false
}
