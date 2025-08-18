package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"time"
	"wave"
	"wave/instruments"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/pixel"
	"github.com/gopxl/pixel/pixelgl"
)

func main() {
	pixelgl.Run(run)
}

func run() {
	// Parse CLA
	mode := flag.String("mode", "One of interactive (start piano), play (play an advanced notation file.)", "interactive,play")
	bpm := flag.Float64("bpm", 60, "What bpm should we run at?")
	defaultInstrumentName := flag.String("instrument", "piano", "Which instrument to use by default, from the config file")
	flag.Parse()

	// Setup the speaker, streamer, and persisten waveform, and start playing
	sr := beep.SampleRate(44100)
	globalWaveform := wave.NewAddWaveform(true)
	globalStreamer := &WaveStreamer{wf: wave.NewLowPassWaveform(globalWaveform, 800)}
	speaker.Init(sr, sr.N(time.Second/40))
	go speaker.Play(globalStreamer)

	// Load the instruments
	inss, err := loadInstruments("instruments.json")
	if err != nil {
		panic(err)
	}

	// Mode switch
	if *mode == "interactive" {
		win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
			Bounds:    pixel.R(0, 0, 50, 50),
			Resizable: true,
		})
		if err != nil {
			panic(err)
		}

		keyMapping := map[pixelgl.Button]struct {
			Key          wave.Key
			OctaveOffset int
		}{
			// Number row (octave offset +1)
			pixelgl.Key1:     {wave.C, 1},
			pixelgl.Key2:     {wave.CSharp, 1},
			pixelgl.Key3:     {wave.D, 1},
			pixelgl.Key4:     {wave.DSharp, 1},
			pixelgl.Key5:     {wave.E, 1},
			pixelgl.Key6:     {wave.F, 1},
			pixelgl.Key7:     {wave.FSharp, 1},
			pixelgl.Key8:     {wave.G, 1},
			pixelgl.Key9:     {wave.GSharp, 1},
			pixelgl.Key0:     {wave.A, 1},
			pixelgl.KeyMinus: {wave.ASharp, 1},
			pixelgl.KeyEqual: {wave.B, 1},

			// QWERTY row (octave offset 0)
			pixelgl.KeyQ:            {wave.C, 0},
			pixelgl.KeyW:            {wave.CSharp, 0},
			pixelgl.KeyE:            {wave.D, 0},
			pixelgl.KeyR:            {wave.DSharp, 0},
			pixelgl.KeyT:            {wave.E, 0},
			pixelgl.KeyY:            {wave.F, 0},
			pixelgl.KeyU:            {wave.FSharp, 0},
			pixelgl.KeyI:            {wave.G, 0},
			pixelgl.KeyO:            {wave.GSharp, 0},
			pixelgl.KeyP:            {wave.A, 0},
			pixelgl.KeyLeftBracket:  {wave.ASharp, 0},
			pixelgl.KeyRightBracket: {wave.B, 0},
		}
		currentKeys := make(map[pixelgl.Button]wave.Waveform)
		currentOctave := 3

		timeline := []timelineEntry{}
		timelineStart := time.Now()
		timeOfLastClick := timelineStart

		for !win.Closed() {
			win.Update()
			// Update synth settings
			if win.JustPressed(pixelgl.KeyZ) {
				currentOctave--
			}
			if win.JustPressed(pixelgl.KeyX) {
				currentOctave++
			}

			if win.JustPressed(pixelgl.KeyLeftShift) || win.JustPressed(pixelgl.KeyLeftControl) {
				timelineStart = time.Now()
				timeOfLastClick = timelineStart
				actions := buildActionsFromTimeline(timeline)
				if win.JustPressed(pixelgl.KeyLeftControl) {
					for i, a := range actions {
						switch a := a.(type) {
						case Advance:
							duration := math.Round(a.Duration*4) / 4
							actions[i] = Advance{Duration: duration}
						}
					}
				}
				fmt.Printf("\n\n\n")
				for _, a := range actions {
					switch a := a.(type) {
					case Play:
						fmt.Printf("%v%v-%v;", a.Key, a.Octave, a.Duration)
					case Advance:
						if a.Duration == 0 {
							continue
						}
						fmt.Printf("%v+;\n", a.Duration)
					}
				}
				timeline = make([]timelineEntry, 0)
			}
			// Update notes
			deleteKeys := make([]pixelgl.Button, 0)
			for btn, tone := range keyMapping {
				currentWaveform, hasCurrentWaveform := currentKeys[btn]
				if win.JustPressed(btn) {
					if hasCurrentWaveform {
						currentWaveform.OffNow()
					}
					ins, ok := inss[*defaultInstrumentName]
					if !ok {
						panic(fmt.Errorf("cannot find instrument '%s'", *defaultInstrumentName))
					}
					wf := ins.CreateWaveform(tone.Key, currentOctave+tone.OctaveOffset, 0)
					currentKeys[btn] = wf
					globalWaveform.Add(currentKeys[btn])
					fmt.Println(tone.Key)
					timeline = append(timeline, timelineEntry{time: time.Since(timelineStart).Seconds(), play: Play{Key: tone.Key, Octave: currentOctave + tone.OctaveOffset, Duration: 10}})
				}
				if win.JustReleased(btn) && hasCurrentWaveform {
					currentWaveform.OffNow()
					deleteKeys = append(deleteKeys, btn)
				}
			}
			for _, k := range deleteKeys {
				delete(currentKeys, k)
			}
			if time.Since(timeOfLastClick).Seconds() > 1/(*bpm/60.0) {
				timeOfLastClick = time.Now()
				globalWaveform.Add(wave.NewRollOffWaveform(
					wave.NewSquareWaveform(50, 0.15, 0),
					0.01,
					1,
				))
			}
		}
	} else if *mode == "play" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		}
		tune, err := ParseAdvancedNotation(string(data), *bpm)
		if err != nil {
			panic(err)
		}
		for {
			for _, action := range tune {
				switch action := action.(type) {
				case Play:
					instrument, ok := inss[action.Instrument]
					if !ok {
						panic(fmt.Errorf("could not find instrument '%s'", action.Instrument))
					}
					globalWaveform.Add(
						instrument.CreateWaveform(action.Key, action.Octave, action.Duration),
					)
					fmt.Println("Play", action.Key, action.Octave, action.Duration)
				case Advance:
					time.Sleep(time.Millisecond * time.Duration(action.Duration*1000))
					fmt.Println("Sleep", action.Duration)
				}
			}
		}
	}
}

func loadInstruments(instrumentConfig string) (map[string]instruments.Instrument, error) {
	f, err := os.Open(instrumentConfig)
	if err != nil {
		return nil, errors.Join(errors.New("could not read instruments config file"), err)
	}
	defer f.Close()
	constructors := make(map[string]instruments.ConstructorJson)
	err = json.NewDecoder(f).Decode(&constructors)
	if err != nil {
		return nil, errors.Join(errors.New("could not parse instruments config file"), err)
	}
	inss := make(map[string]instruments.Instrument)
	for name, cons := range constructors {
		ins, err := instruments.NewInstrument(cons)
		if ins == nil {
			panic("a")
		}
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to construct instrument '%s'", name), err)
		}
		inss[name] = ins
	}
	return inss, nil
}
