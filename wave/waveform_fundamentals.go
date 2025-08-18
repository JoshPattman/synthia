package wave

import "math"

func NewSinWaveform(freq, vol float64) Waveform {
	return &sinWaveform{f: freq, v: vol}
}
func NewSquareWaveform(freq, vol, thresh float64) Waveform {
	return &squareWaveform{f: freq, v: vol, threshold: thresh, minIsZero: false}
}
func NewZeroSquareWaveform(freq, vol, thresh float64) Waveform {
	return &squareWaveform{f: freq, v: vol, threshold: thresh, minIsZero: true}
}
func NewTriangleWaveform(freq, vol float64) Waveform {
	return &triangleWaveform{f: freq, v: vol}
}
func NewSawtoothWaveform(freq, vol float64) Waveform {
	return &sawWaveform{f: freq, v: vol}
}

// ----------------- Sine -----------------
type sinWaveform struct {
	f float64
	v float64
	t float64
}

func (s *sinWaveform) OffNow() {
	s.v = 0
}

func (s *sinWaveform) Next(deltaTime float64) (float64, bool) {
	if s.v == 0 {
		return 0, true
	}
	s.t += deltaTime
	return math.Sin(s.t*math.Pi*2*s.f) * s.v, false
}

// ----------------- Square -----------------
type squareWaveform struct {
	f         float64
	v         float64
	threshold float64 // between -1 and 1
	t         float64
	minIsZero bool
}

// OffNow implements Waveform.
func (s *squareWaveform) OffNow() {
	s.v = 0
}

func (s *squareWaveform) Next(deltaTime float64) (float64, bool) {
	if s.v == 0 {
		return 0, true
	}
	s.t += deltaTime
	if math.Sin(s.t*math.Pi*2*s.f) >= s.threshold {
		return s.v, false
	}
	if s.minIsZero {
		return 0, false
	} else {
		return -s.v, false
	}
}

// ----------------- Triangle -----------------
type triangleWaveform struct {
	f float64
	v float64
	t float64
}

// OffNow implements Waveform.
func (t *triangleWaveform) OffNow() {
	t.v = 0
}

func (t *triangleWaveform) Next(deltaTime float64) (float64, bool) {
	if t.v == 0 {
		return 0, true
	}
	t.t += deltaTime
	phase := t.t * t.f
	frac := phase - math.Floor(phase) // 0..1
	val := 4*math.Abs(frac-0.5) - 1   // -1..1
	return val * t.v, false
}

// ----------------- Sawtooth -----------------
type sawWaveform struct {
	f float64
	v float64
	t float64
}

// OffNow implements Waveform.
func (s *sawWaveform) OffNow() {
	s.v = 0
}

func (s *sawWaveform) Next(deltaTime float64) (float64, bool) {
	if s.v == 0 {
		return 0, true
	}
	s.t += deltaTime
	phase := s.t * s.f
	frac := phase - math.Floor(phase) // 0..1
	val := 2*frac - 1                 // -1..1
	return val * s.v, false
}
