package wave

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

type Key uint8

const (
	C Key = iota
	CSharp
	D
	DSharp
	E
	F
	FSharp
	G
	GSharp
	A
	ASharp
	B
)

func ParseStringToKey(s string) (Key, error) {
	if k, ok := keyFromString[strings.ToUpper(s)]; ok {
		return k, nil
	} else {
		return 0, fmt.Errorf("not valid string key")
	}
}

var keyFromString = map[string]Key{
	"C":  C,
	"C#": CSharp,
	"D":  D,
	"D#": DSharp,
	"E":  E,
	"F":  F,
	"F#": FSharp,
	"G":  G,
	"G#": GSharp,
	"A":  A,
	"A#": ASharp,
	"B":  B,
}

var keyToString = map[Key]string{
	C:      "C",
	CSharp: "C#",
	D:      "D",
	DSharp: "D#",
	E:      "E",
	F:      "F",
	FSharp: "F#",
	G:      "G",
	GSharp: "G#",
	A:      "A",
	ASharp: "A#",
	B:      "B",
}

func (k Key) String() string {
	return keyToString[k]
}

func (k *Key) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, ok := keyFromString[s]
	if !ok {
		return fmt.Errorf("invalid key: %s", s)
	}
	*k = val
	return nil
}

func (k Key) MarshalJSON() ([]byte, error) {
	s, ok := keyToString[k]
	if !ok {
		return nil, fmt.Errorf("invalid key value: %d", k)
	}
	return json.Marshal(s)
}

func NoteToFreq(note Key, octave int) float64 {
	index := -1
	switch note {
	case C:
		index = 0
	case CSharp:
		index = 1
	case D:
		index = 2
	case DSharp:
		index = 3
	case E:
		index = 4
	case F:
		index = 5
	case FSharp:
		index = 6
	case G:
		index = 7
	case GSharp:
		index = 8
	case A:
		index = 9
	case ASharp:
		index = 10
	case B:
		index = 11
	default:
		panic("Not a valid note type")
	}
	n := (octave-4)*12 + (index - 9)
	return 440 * math.Exp2(float64(n)/12)
}

func SpeedChange(fromKey Key, fromOctave int, toKey Key, toOctave int) float64 {
	f1 := NoteToFreq(fromKey, fromOctave)
	f2 := NoteToFreq(toKey, toOctave)
	return f2 / f1
}

func TryMakeSharp(key Key) (Key, bool) {
	switch key {
	case C:
		return CSharp, true
	case D:
		return DSharp, true
	case F:
		return FSharp, true
	case G:
		return GSharp, true
	case A:
		return ASharp, true
	default:
		return key, false
	}
}
