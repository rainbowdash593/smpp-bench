package statistics

import (
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	Sent            int64
	SentLatencyMs   int64
	Failed          int64
	FailedLatencyMs int64
	ErrorRate       float64
	DLR             int64
}

type TotalStats struct {
	Stats
	MaxRPS float64
}
type TickStats struct {
	Stats
	Timestamp time.Time
	RPS       float64
}

type Collector struct {
	sent                 atomic.Int64
	sentTotalLatencyMs   atomic.Int64
	failed               atomic.Int64
	failedTotalLatencyMs atomic.Int64
	dlr                  atomic.Int64

	mu     sync.Mutex
	maxRps float64

	lastTick time.Time

	tickSent                 atomic.Int64
	tickSentTotalLatencyMs   atomic.Int64
	tickFailed               atomic.Int64
	tickFailedTotalLatencyMs atomic.Int64
	tickDLR                  atomic.Int64
}

func NewCollector() *Collector {
	return &Collector{
		mu:       sync.Mutex{},
		lastTick: time.Now(),
	}
}

func (c *Collector) RecordSuccess(latency int64) {
	c.sent.Add(1)
	c.sentTotalLatencyMs.Add(latency)

	c.tickSent.Add(1)
	c.tickSentTotalLatencyMs.Add(latency)
}

func (c *Collector) RecordFailed(latency int64) {
	c.failed.Add(1)
	c.failedTotalLatencyMs.Add(latency)

	c.tickFailed.Add(1)
	c.tickFailedTotalLatencyMs.Add(latency)
}

func (c *Collector) RecordDLR() {
	c.dlr.Add(1)
	c.tickDLR.Add(1)
}

func (c *Collector) Tick() TickStats {
	sent := c.tickSent.Swap(0)
	failed := c.tickFailed.Swap(0)
	totalSentLatency := c.tickSentTotalLatencyMs.Swap(0)
	totalFailedLatency := c.tickFailedTotalLatencyMs.Swap(0)
	dlr := c.tickDLR.Swap(0)

	c.mu.Lock()
	now := time.Now()
	elapsed := now.Sub(c.lastTick).Seconds()
	c.lastTick = now
	c.mu.Unlock()

	var (
		sentLatency   int64
		failedLatency int64
		errRate       float64
		rps           float64
	)

	if sent > 0 {
		sentLatency = totalSentLatency / sent
	}
	if failed > 0 {
		failedLatency = totalFailedLatency / failed
	}
	total := sent + failed
	if total > 0 {
		errRate = float64(failed) / float64(total) * 100
	}
	if elapsed > 0 {
		rps = float64(sent) / elapsed
	}

	c.mu.Lock()
	if rps > c.maxRps {
		c.maxRps = rps
	}
	c.mu.Unlock()

	return TickStats{
		Stats: Stats{
			Sent:            sent,
			SentLatencyMs:   sentLatency,
			Failed:          failed,
			FailedLatencyMs: failedLatency,
			ErrorRate:       errRate,
			DLR:             dlr,
		},
		Timestamp: now,
		RPS:       rps,
	}
}

func (c *Collector) Total() TotalStats {
	sent := c.sent.Swap(0)
	failed := c.failed.Swap(0)
	totalSentLatency := c.sentTotalLatencyMs.Swap(0)
	totalFailedLatency := c.failedTotalLatencyMs.Swap(0)
	dlr := c.dlr.Swap(0)
	var (
		sentLatency   int64
		failedLatency int64
		errRate       float64
	)

	if sent > 0 {
		sentLatency = totalSentLatency / sent
	}
	if failed > 0 {
		failedLatency = totalFailedLatency / failed
	}
	total := sent + failed
	if total > 0 {
		errRate = float64(failed) / float64(total) * 100
	}

	return TotalStats{
		Stats: Stats{
			Sent:            sent,
			SentLatencyMs:   sentLatency,
			Failed:          failed,
			FailedLatencyMs: failedLatency,
			ErrorRate:       errRate,
			DLR:             dlr,
		},
		MaxRPS: c.maxRps,
	}
}
