package player

import (
	"math"
	"sync"

	"github.com/gopxl/beep/v2"
)

const (
	// tapBufSize is the number of stereo samples kept in the ring buffer.
	// At 44100Hz this is ~93ms of audio.
	tapBufSize = 4096

	// NumBands is the number of frequency bands for the visualizer.
	NumBands = 16
)

// Tap is a beep.Streamer that passes audio through while copying samples
// to a ring buffer for visualization.
type Tap struct {
	Streamer beep.Streamer

	mu  sync.Mutex
	buf [tapBufSize][2]float64
	pos int // write position in the ring buffer

	// Smoothed band values for visual continuity
	smooth [NumBands]float64
}

// Stream implements beep.Streamer. It streams from the wrapped streamer
// and copies samples to the ring buffer.
func (t *Tap) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = t.Streamer.Stream(samples)
	if n > 0 {
		t.mu.Lock()
		for i := 0; i < n; i++ {
			t.buf[t.pos] = samples[i]
			t.pos = (t.pos + 1) % tapBufSize
		}
		t.mu.Unlock()
	}
	return n, ok
}

// Err implements beep.Streamer.
func (t *Tap) Err() error {
	return t.Streamer.Err()
}

// Bands returns the amplitude for each frequency band, suitable for driving
// a bar visualizer. Values are in the range [0, 1].
//
// Uses a simple DFT on selected frequency bins to separate bass from treble,
// then applies smoothing for visual appeal.
func (t *Tap) Bands() [NumBands]float64 {
	t.mu.Lock()
	// Copy the most recent window of samples (last 1024).
	const windowSize = 1024
	var window [windowSize]float64
	for i := 0; i < windowSize; i++ {
		idx := (t.pos - windowSize + i + tapBufSize) % tapBufSize
		// Mix stereo to mono
		window[i] = (t.buf[idx][0] + t.buf[idx][1]) * 0.5
	}
	t.mu.Unlock()

	// Apply Hann window to reduce spectral leakage
	for i := range window {
		w := 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(i)/float64(windowSize-1)))
		window[i] *= w
	}

	// Compute magnitude for selected frequency bins using DFT.
	// Map 16 bands to roughly log-spaced frequency ranges.
	// At 44100Hz with 1024 samples, each bin = ~43Hz.
	// Band 0 = ~43-86Hz (bass), Band 15 = ~8k-16kHz (treble)
	binStarts := [NumBands]int{1, 2, 3, 4, 6, 8, 11, 15, 20, 27, 36, 48, 64, 85, 113, 150}
	binEnds := [NumBands]int{2, 3, 4, 6, 8, 11, 15, 20, 27, 36, 48, 64, 85, 113, 150, 200}

	var bands [NumBands]float64
	for b := 0; b < NumBands; b++ {
		var maxMag float64
		for k := binStarts[b]; k < binEnds[b] && k < windowSize/2; k++ {
			// DFT for bin k: X[k] = sum(x[n] * e^(-j*2*pi*k*n/N))
			var re, im float64
			for n := 0; n < windowSize; n++ {
				angle := 2.0 * math.Pi * float64(k) * float64(n) / float64(windowSize)
				re += window[n] * math.Cos(angle)
				im -= window[n] * math.Sin(angle)
			}
			mag := math.Sqrt(re*re+im*im) / float64(windowSize)
			if mag > maxMag {
				maxMag = mag
			}
		}
		// Scale with aggressive gain + compression
		scaled := maxMag * 25.0
		if scaled > 0 {
			scaled = math.Sqrt(scaled)
		}
		bands[b] = clamp(scaled, 0, 1)
	}

	// Smooth with previous frame for visual continuity (decay + attack)
	const attack = 0.6  // how fast bars rise
	const decay = 0.15  // how fast bars fall
	for i := range bands {
		if bands[i] > t.smooth[i] {
			t.smooth[i] += (bands[i] - t.smooth[i]) * attack
		} else {
			t.smooth[i] += (bands[i] - t.smooth[i]) * decay
		}
		bands[i] = t.smooth[i]
	}

	return bands
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
