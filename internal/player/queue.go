package player

import (
	"math/rand/v2"
	"sync"
)

type QueueMode int

const (
	QueueNormal    QueueMode = iota
	QueueShuffle
	QueueRepeatOne
)

type Queue struct {
	mu           sync.Mutex
	tracks       []Track
	index        int
	mode         QueueMode
	shuffleOrder []int
}

func NewQueue() *Queue {
	return &Queue{}
}

// SetTracks replaces the queue with the given tracks, starting at startIndex.
func (q *Queue) SetTracks(tracks []Track, startIndex int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = make([]Track, len(tracks))
	copy(q.tracks, tracks)
	q.index = startIndex
	if q.mode == QueueShuffle {
		q.buildShuffleOrder()
	}
}

// Current returns the track at the current position, or nil if empty.
func (q *Queue) Current() *Track {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.currentLocked()
}

func (q *Queue) currentLocked() *Track {
	if len(q.tracks) == 0 {
		return nil
	}
	idx := q.resolveIndex()
	if idx < 0 || idx >= len(q.tracks) {
		return nil
	}
	t := q.tracks[idx]
	return &t
}

// Next advances to the next track and returns it, or nil if at the end.
func (q *Queue) Next() *Track {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.tracks) == 0 {
		return nil
	}

	switch q.mode {
	case QueueRepeatOne:
		// Stay on the same track
		return q.currentLocked()
	case QueueShuffle:
		q.index++
		if q.index >= len(q.shuffleOrder) {
			return nil
		}
		return q.currentLocked()
	default:
		q.index++
		if q.index >= len(q.tracks) {
			return nil
		}
		return q.currentLocked()
	}
}

// Previous goes back one track and returns it, or nil if at the start.
func (q *Queue) Previous() *Track {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.tracks) == 0 {
		return nil
	}

	switch q.mode {
	case QueueRepeatOne:
		return q.currentLocked()
	case QueueShuffle:
		if q.index > 0 {
			q.index--
		}
		return q.currentLocked()
	default:
		if q.index > 0 {
			q.index--
		}
		return q.currentLocked()
	}
}

// HasNext returns true if there is a next track.
func (q *Queue) HasNext() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.mode == QueueRepeatOne {
		return len(q.tracks) > 0
	}
	if q.mode == QueueShuffle {
		return q.index < len(q.shuffleOrder)-1
	}
	return q.index < len(q.tracks)-1
}

// HasPrevious returns true if there is a previous track.
func (q *Queue) HasPrevious() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.index > 0
}

// SetMode changes the queue mode.
func (q *Queue) SetMode(mode QueueMode) {
	q.mu.Lock()
	defer q.mu.Unlock()

	oldMode := q.mode
	q.mode = mode

	if mode == QueueShuffle && oldMode != QueueShuffle {
		q.buildShuffleOrder()
	}
	if mode != QueueShuffle && oldMode == QueueShuffle {
		// Snap back: find the real index of the current track
		if len(q.shuffleOrder) > 0 && q.index < len(q.shuffleOrder) {
			q.index = q.shuffleOrder[q.index]
		}
		q.shuffleOrder = nil
	}
}

// Mode returns the current queue mode.
func (q *Queue) Mode() QueueMode {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.mode
}

// Clear empties the queue.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = nil
	q.index = 0
	q.shuffleOrder = nil
}

// Len returns the number of tracks in the queue.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.tracks)
}

// Index returns the current position (1-based for display).
func (q *Queue) Index() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.index + 1
}

// resolveIndex returns the actual track index, accounting for shuffle.
func (q *Queue) resolveIndex() int {
	if q.mode == QueueShuffle && len(q.shuffleOrder) > 0 {
		if q.index < len(q.shuffleOrder) {
			return q.shuffleOrder[q.index]
		}
		return -1
	}
	return q.index
}

// buildShuffleOrder creates a random permutation with the current track first.
func (q *Queue) buildShuffleOrder() {
	n := len(q.tracks)
	if n == 0 {
		q.shuffleOrder = nil
		return
	}

	currentReal := q.index
	if currentReal >= n {
		currentReal = 0
	}

	// Build permutation of all indices except current
	order := make([]int, 0, n-1)
	for i := 0; i < n; i++ {
		if i != currentReal {
			order = append(order, i)
		}
	}
	rand.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})

	// Current track goes first
	q.shuffleOrder = make([]int, 0, n)
	q.shuffleOrder = append(q.shuffleOrder, currentReal)
	q.shuffleOrder = append(q.shuffleOrder, order...)
	q.index = 0
}
