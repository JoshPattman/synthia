package wave

type Waveform interface {
	// What is the point on the wave after this time.
	// Also return a value, which when true, means the note can safely be ignored from now on.
	Next(deltaTime float64) (float64, bool)
	OffNow()
}
