package util

import (
	"time"
)

// Unix timestamps in seconds (blockchain standard)

// Now returns the current Unix timestamp in seconds
func Now() uint64 {
	return uint64(time.Now().Unix())
}

// NowMilli returns the current Unix timestamp in milliseconds
func NowMilli() uint64 {
	return uint64(time.Now().UnixMilli())
}

// NowNano returns the current Unix timestamp in nanoseconds
func NowNano() uint64 {
	return uint64(time.Now().UnixNano())
}

// FromUnix converts Unix seconds to time.Time
func FromUnix(ts uint64) time.Time {
	return time.Unix(int64(ts), 0)
}

// FromUnixMilli converts Unix milliseconds to time.Time
func FromUnixMilli(ts uint64) time.Time {
	return time.UnixMilli(int64(ts))
}

// FormatTime formats a Unix timestamp as RFC3339
func FormatTime(ts uint64) string {
	return FromUnix(ts).Format(time.RFC3339)
}

// FormatDuration formats a duration in human-readable form
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return d.String()
	}
	if d < time.Minute {
		return d.Round(time.Millisecond).String()
	}
	if d < time.Hour {
		return d.Round(time.Second).String()
	}
	return d.Round(time.Minute).String()
}

// ParseDuration parses a duration string
func ParseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// BlockTime calculates when a block should be produced
func BlockTime(genesisTime uint64, blockNumber uint64, blockInterval uint64) uint64 {
	return genesisTime + (blockNumber * blockInterval)
}

// SlotNumber calculates the current slot number
func SlotNumber(genesisTime uint64, slotDuration uint64) uint64 {
	if Now() < genesisTime {
		return 0
	}
	return (Now() - genesisTime) / slotDuration
}

// SlotStartTime returns the start time of a slot
func SlotStartTime(genesisTime uint64, slot uint64, slotDuration uint64) uint64 {
	return genesisTime + (slot * slotDuration)
}

// TimeUntilSlot returns duration until a specific slot
func TimeUntilSlot(genesisTime uint64, slot uint64, slotDuration uint64) time.Duration {
	slotStart := SlotStartTime(genesisTime, slot, slotDuration)
	now := Now()
	if slotStart <= now {
		return 0
	}
	return time.Duration(slotStart-now) * time.Second
}

// EpochNumber calculates the current epoch
func EpochNumber(genesisTime uint64, epochDuration uint64) uint64 {
	if Now() < genesisTime {
		return 0
	}
	return (Now() - genesisTime) / epochDuration
}

// IsValidTimestamp checks if a timestamp is within acceptable range
func IsValidTimestamp(ts uint64, maxDrift time.Duration) bool {
	now := Now()
	drift := uint64(maxDrift.Seconds())
	
	// Not too far in the future
	if ts > now+drift {
		return false
	}
	
	// Not too far in the past (optional, depends on use case)
	return true
}

// Timer utilities for block production

// BlockTimer manages block production timing
type BlockTimer struct {
	genesisTime   uint64
	blockInterval uint64
	ticker        *time.Ticker
	stop          chan struct{}
}

// NewBlockTimer creates a new block timer
func NewBlockTimer(genesisTime uint64, blockInterval uint64) *BlockTimer {
	return &BlockTimer{
		genesisTime:   genesisTime,
		blockInterval: blockInterval,
		stop:          make(chan struct{}),
	}
}

// Start starts the block timer
func (bt *BlockTimer) Start(callback func(slot uint64)) {
	bt.ticker = time.NewTicker(time.Duration(bt.blockInterval) * time.Second)
	
	go func() {
		for {
			select {
			case <-bt.ticker.C:
				slot := SlotNumber(bt.genesisTime, bt.blockInterval)
				callback(slot)
			case <-bt.stop:
				return
			}
		}
	}()
}

// Stop stops the block timer
func (bt *BlockTimer) Stop() {
	close(bt.stop)
	if bt.ticker != nil {
		bt.ticker.Stop()
	}
}

// WaitForNextSlot waits until the next slot
func (bt *BlockTimer) WaitForNextSlot() uint64 {
	currentSlot := SlotNumber(bt.genesisTime, bt.blockInterval)
	nextSlot := currentSlot + 1
	duration := TimeUntilSlot(bt.genesisTime, nextSlot, bt.blockInterval)
	time.Sleep(duration)
	return nextSlot
}
