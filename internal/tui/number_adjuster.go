package tui

import "time"

type NumberAdjuster struct {
	repeatCount   int
	lastKeyTime   time.Time
	current       float64
	max           float64
	min           float64
	baseIncrement float64
}

func NewNumberAdjuster(current, max, min, baseIncrement float64) *NumberAdjuster {
	return &NumberAdjuster{
		current:       current,
		max:           max,
		min:           min,
		baseIncrement: baseIncrement,
		repeatCount:   0,
	}
}

func (s *NumberAdjuster) Value() float64 {
	return s.current
}

func (s *NumberAdjuster) SetCurrent(current float64) {
	s.current = current
}

func (s *NumberAdjuster) Increase() {
	s.adjust(1.0)
}

func (s *NumberAdjuster) Decrease() {
	s.adjust(-1.0)
}

func (s *NumberAdjuster) adjust(mult float64) {
	now := time.Now()

	if now.Sub(s.lastKeyTime) < 200*time.Millisecond {
		s.repeatCount++
	} else {
		s.repeatCount = 0
	}

	s.lastKeyTime = now

	multiplier := 1.0
	switch {
	case s.repeatCount > 5:
		multiplier = 10.0
	case s.repeatCount > 3:
		multiplier = 5.0
	case s.repeatCount > 1:
		multiplier = 2.0
	}

	increment := mult * s.baseIncrement * multiplier
	current := s.current + increment

	if current < s.min {
		current = s.min
	}
	if current > s.max {
		current = s.max
	}

	s.current = current
}
