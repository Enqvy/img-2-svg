package main

import (
	"fmt"
	"time"
)

type ProgressTracker struct {
	total      int
	processed  int
	quiet      bool
	startTime  time.Time
	lastUpdate time.Time
}

func NewProgressTracker(total int, quiet bool) *ProgressTracker {
	return &ProgressTracker{
		total:      total,
		quiet:      quiet,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}
}

func (p *ProgressTracker) Update(increment int) {
	if p.quiet {
		return
	}

	p.processed += increment

	if time.Since(p.lastUpdate) < 100*time.Millisecond && p.processed < p.total {
		return
	}
	p.lastUpdate = time.Now()

	percent := float64(p.processed) / float64(p.total) * 100
	bar := p.createProgressBar(percent)

	elapsed := time.Since(p.startTime)
	eta := p.calculateETA(percent, elapsed)

	fmt.Printf("\r%s %.1f%% ETA: %v", bar, percent, eta.Round(time.Second))
}

func (p *ProgressTracker) Finish() {
	if p.quiet {
		return
	}
	fmt.Printf("\r[==================================================] 100.0%% ETA: 0s\n")
}

func (p *ProgressTracker) createProgressBar(percent float64) string {
	const barWidth = 50
	completed := int(float64(barWidth) * percent / 100)

	bar := "["
	for i := 0; i < barWidth; i++ {
		switch {
		case i < completed:
			bar += "="
		case i == completed:
			bar += ">"
		default:
			bar += " "
		}
	}
	bar += "]"

	return bar
}

func (p *ProgressTracker) calculateETA(percent float64, elapsed time.Duration) time.Duration {
	if percent == 0 {
		return 0
	}
	totalEstimate := time.Duration(float64(elapsed) / percent * 100)
	return totalEstimate - elapsed
}