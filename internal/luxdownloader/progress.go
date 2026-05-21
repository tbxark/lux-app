package luxdownloader

import (
	"io"
	"sync"
	"time"
)

const defaultProgressThrottle = 100 * time.Millisecond

type progressTracker struct {
	mu       sync.Mutex
	callback ProgressFunc
	throttle time.Duration
	lastEmit time.Time
	event    ProgressEvent
}

func newProgressTracker(event ProgressEvent, callback ProgressFunc, throttle time.Duration) *progressTracker {
	if throttle <= 0 {
		throttle = defaultProgressThrottle
	}
	if event.Total > 0 && event.Current > event.Total {
		event.Current = event.Total
	}
	event.Percent = percent(event.Current, event.Total)
	return &progressTracker{
		callback: callback,
		throttle: throttle,
		event:    event,
	}
}

func (tracker *progressTracker) emit(force bool) {
	if tracker == nil || tracker.callback == nil {
		return
	}
	now := time.Now()
	tracker.mu.Lock()
	if !force && !tracker.lastEmit.IsZero() && now.Sub(tracker.lastEmit) < tracker.throttle {
		tracker.mu.Unlock()
		return
	}
	tracker.lastEmit = now
	event := tracker.event
	tracker.mu.Unlock()
	tracker.callback(event)
}

func (tracker *progressTracker) add(delta int64) {
	if tracker == nil || delta <= 0 {
		return
	}
	tracker.mu.Lock()
	tracker.event.Current += delta
	if tracker.event.Total > 0 && tracker.event.Current > tracker.event.Total {
		tracker.event.Current = tracker.event.Total
	}
	tracker.event.Percent = percent(tracker.event.Current, tracker.event.Total)
	tracker.mu.Unlock()
	tracker.emit(false)
}

func (tracker *progressTracker) setPhase(phase ProgressPhase, message string, force bool) {
	if tracker == nil {
		return
	}
	tracker.mu.Lock()
	tracker.event.Phase = phase
	tracker.event.Message = message
	tracker.event.Percent = percent(tracker.event.Current, tracker.event.Total)
	tracker.mu.Unlock()
	tracker.emit(force)
}

func (tracker *progressTracker) complete(phase ProgressPhase, message string) {
	if tracker == nil {
		return
	}
	tracker.mu.Lock()
	tracker.event.Phase = phase
	tracker.event.Message = message
	if tracker.event.Total > 0 {
		tracker.event.Current = tracker.event.Total
	}
	tracker.event.Percent = percent(tracker.event.Current, tracker.event.Total)
	tracker.mu.Unlock()
	tracker.emit(true)
}

func percent(current, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(current) / float64(total)
}

type progressWriter struct {
	writer  io.Writer
	tracker *progressTracker
}

func (writer *progressWriter) Write(p []byte) (int, error) {
	n, err := writer.writer.Write(p)
	if n > 0 {
		writer.tracker.add(int64(n))
	}
	return n, err
}
