package player

import (
	"fmt"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/speaker"
)

// Track holds metadata about a track.
type Track struct {
	ID       int64
	Title    string
	Artist   string
	Album    string
	Duration float64
}

// State represents the player's current state.
type State int

const (
	Stopped State = iota
	Playing
	Paused
)

// PlayerEvent identifies what happened.
type PlayerEvent int

const (
	EventStopped PlayerEvent = iota
	EventAdvance
)

// PlayerNotification is sent on the events channel.
type PlayerNotification struct {
	Event PlayerEvent
	Track Track // set for EventAdvance
}

const (
	speakerSampleRate = 44100
	speakerBufSize    = 2048 // ~46ms at 44100Hz -- low latency
)

// Player manages audio playback via the beep audio pipeline.
type Player struct {
	mu         sync.Mutex
	state      State
	current    *Track
	queue      *Queue
	manualStop bool
	generation int // incremented on each Play() to invalidate old goroutines

	// beep pipeline components
	streamer beep.StreamSeekCloser // decoded audio source
	ctrl     *beep.Ctrl           // pause/resume control
	volume   *effects.Volume      // volume control
	tap      *Tap                 // PCM sample tap for visualizer
	seekable bool                 // whether the current stream supports seeking
	format   beep.Format          // source format (for position calculations)

	speakerReady bool

	// Event channel for TUI communication
	events chan PlayerNotification
}

func New() *Player {
	return &Player{
		queue:  NewQueue(),
		events: make(chan PlayerNotification, 1),
	}
}

// Queue returns the player's queue.
func (p *Player) Queue() *Queue {
	return p.queue
}

// Tap returns the audio tap for visualization, or nil if not playing.
func (p *Player) Tap() *Tap {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.tap
}

// initSpeaker initializes the speaker on first use.
func (p *Player) initSpeaker() error {
	if p.speakerReady {
		return nil
	}
	err := speaker.Init(beep.SampleRate(speakerSampleRate), speakerBufSize)
	if err != nil {
		return fmt.Errorf("initializing speaker: %w", err)
	}
	p.speakerReady = true
	return nil
}

// Play starts playing the given URL. Stops any currently playing track first.
func (p *Player) Play(url string, track Track) error {
	p.stopCurrent()

	// Fetch and decode the audio stream.
	stream, err := openAudioStream(url)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.initSpeaker(); err != nil {
		stream.Streamer.Close()
		return err
	}

	// Build the pipeline: decoder -> resampler -> volume -> tap -> ctrl -> speaker
	var src beep.Streamer = stream.Streamer
	if stream.Format.SampleRate != beep.SampleRate(speakerSampleRate) {
		src = beep.Resample(4, stream.Format.SampleRate, beep.SampleRate(speakerSampleRate), src)
	}

	vol := &effects.Volume{Streamer: src, Base: 2, Volume: 0, Silent: false}
	tap := &Tap{Streamer: vol}
	ctrl := &beep.Ctrl{Streamer: tap, Paused: false}

	p.streamer = stream.Streamer
	p.format = stream.Format
	p.seekable = stream.Seekable
	p.volume = vol
	p.tap = tap
	p.ctrl = ctrl
	p.state = Playing
	p.current = &track
	p.manualStop = false
	p.generation++
	gen := p.generation

	// Play through the speaker. When the stream ends, the callback fires.
	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		p.mu.Lock()
		// If generation changed, a new Play() already took over -- bail.
		if p.generation != gen {
			p.mu.Unlock()
			return
		}

		wasManualStop := p.manualStop
		p.state = Stopped
		p.clearPipelineLocked()

		if wasManualStop {
			p.current = nil
			p.mu.Unlock()
			return
		}

		// Track ended naturally -- try to advance.
		next := p.queue.Next()
		if next == nil {
			p.current = nil
			p.mu.Unlock()
			select {
			case <-p.events:
			default:
			}
			p.events <- PlayerNotification{Event: EventStopped}
			return
		}

		t := *next
		p.current = nil
		p.mu.Unlock()

		select {
		case <-p.events:
		default:
		}
		p.events <- PlayerNotification{Event: EventAdvance, Track: t}
	})))

	return nil
}

// clearPipelineLocked resets pipeline pointers. Must be called with mu held.
func (p *Player) clearPipelineLocked() {
	if p.streamer != nil {
		p.streamer.Close()
	}
	p.streamer = nil
	p.ctrl = nil
	p.volume = nil
	p.tap = nil
}

// stopCurrent stops the current playback and clears the pipeline.
func (p *Player) stopCurrent() {
	p.mu.Lock()
	if p.state == Stopped {
		p.mu.Unlock()
		return
	}
	p.manualStop = true
	p.state = Stopped
	p.current = nil
	p.generation++ // invalidate the completion callback

	// Clear the speaker and close the stream.
	speaker.Clear()
	p.clearPipelineLocked()
	p.mu.Unlock()
}

// Stop stops playback. Does not auto-advance.
func (p *Player) Stop() {
	p.stopCurrent()
}

// Pause suspends playback.
func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != Playing || p.ctrl == nil {
		return
	}

	speaker.Lock()
	p.ctrl.Paused = true
	speaker.Unlock()

	p.state = Paused
}

// Resume resumes playback.
func (p *Player) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != Paused || p.ctrl == nil {
		return
	}

	speaker.Lock()
	p.ctrl.Paused = false
	speaker.Unlock()

	p.state = Playing
}

// TogglePause toggles between playing and paused.
func (p *Player) TogglePause() {
	s := p.GetState()
	if s == Playing {
		p.Pause()
	} else if s == Paused {
		p.Resume()
	}
}

// Elapsed returns the elapsed playback time using exact sample position.
func (p *Player) Elapsed() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer == nil || p.state == Stopped {
		return 0
	}

	speaker.Lock()
	pos := p.streamer.Position()
	speaker.Unlock()

	return p.format.SampleRate.D(pos)
}

// TotalDuration returns the total duration from the decoded stream.
func (p *Player) TotalDuration() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer == nil {
		return 0
	}

	speaker.Lock()
	length := p.streamer.Len()
	speaker.Unlock()

	return p.format.SampleRate.D(length)
}

// Seek jumps to a position relative to the current position.
func (p *Player) Seek(delta time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer == nil || !p.seekable {
		return
	}

	speaker.Lock()
	pos := p.streamer.Position()
	length := p.streamer.Len()
	newPos := pos + p.format.SampleRate.N(delta)
	if newPos < 0 {
		newPos = 0
	}
	if newPos >= length {
		newPos = length - 1
	}
	_ = p.streamer.Seek(newPos)
	speaker.Unlock()
}

// IsSeekable returns whether the current stream supports seeking.
func (p *Player) IsSeekable() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.seekable
}

// SetVolume sets the volume. 0 is neutral, negative is quieter, positive is louder.
// Range: roughly -5 (very quiet) to +1 (loud).
func (p *Player) SetVolume(v float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.volume == nil {
		return
	}

	speaker.Lock()
	p.volume.Volume = v
	speaker.Unlock()
}

// Volume returns the current volume level.
func (p *Player) Volume() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.volume == nil {
		return 0
	}

	speaker.Lock()
	v := p.volume.Volume
	speaker.Unlock()

	return v
}

// AdjustVolume changes volume by delta (e.g., +0.5 or -0.5).
func (p *Player) AdjustVolume(delta float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.volume == nil {
		return
	}

	speaker.Lock()
	p.volume.Volume = clamp(p.volume.Volume+delta, -5, 1)
	speaker.Unlock()
}

// GetState returns the current playback state.
func (p *Player) GetState() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// Current returns the currently playing track, or nil.
func (p *Player) Current() *Track {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current == nil {
		return nil
	}
	t := *p.current
	return &t
}

// IsPlaying returns true if audio is currently playing or paused.
func (p *Player) IsPlaying() bool {
	s := p.GetState()
	return s == Playing || s == Paused
}

// WaitForEvent blocks until the player sends an event (track ended or advanced).
func (p *Player) WaitForEvent() PlayerNotification {
	return <-p.events
}

// SkipNext stops the current track to trigger auto-advance.
func (p *Player) SkipNext() {
	p.mu.Lock()
	if p.state == Stopped || p.ctrl == nil {
		p.mu.Unlock()
		return
	}
	p.manualStop = false // let auto-advance happen

	// Setting ctrl.Streamer to nil causes Ctrl to report "drained",
	// which makes beep.Seq advance to the completion Callback.
	speaker.Lock()
	p.ctrl.Streamer = nil
	speaker.Unlock()

	// Close the underlying decoder -- the callback will handle the rest.
	if p.streamer != nil {
		p.streamer.Close()
		p.streamer = nil
	}
	p.mu.Unlock()
}

// SkipPrevious goes to the previous track (or restarts current if >3s elapsed).
func (p *Player) SkipPrevious() *Track {
	if p.Elapsed() > 3*time.Second {
		return p.Current()
	}
	p.mu.Lock()
	prev := p.queue.Previous()
	p.mu.Unlock()
	if prev == nil {
		return p.Current()
	}
	return prev
}

// Cleanup releases all audio resources. Call on app exit.
func (p *Player) Cleanup() {
	p.stopCurrent()
	if p.speakerReady {
		speaker.Close()
	}
}
