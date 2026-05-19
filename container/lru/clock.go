package lru

import (
	"time"
)

type clock interface {
	Now() time.Time
	Ticker(dur time.Duration) (ch <-chan time.Time, stop func())
}

type systemClock struct{}

func (s systemClock) Now() time.Time { return time.Now() }

func (s systemClock) Ticker(dur time.Duration) (ch <-chan time.Time, stop func()) {
	ticker := time.NewTicker(dur)

	return ticker.C, ticker.Stop
}
