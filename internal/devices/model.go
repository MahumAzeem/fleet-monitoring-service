package devices

import (
	"fmt"
	"sync"
	"time"
)

type ErrorResponse struct {
	Msg string `json:"msg"`
}

type StatsRequest struct {
	SentAt     time.Time `json:"sent_at"`
	UploadTime int64     `json:"upload_time"`
}

type StatsResponse struct {
	Uptime        float64 `json:"uptime"`
	AvgUploadTime string  `json:"avg_upload_time"`
}

type DeviceStats struct {
	mu sync.RWMutex

	HeartbeatMinutes map[int64]struct{}
	FirstHeartbeat   time.Time
	LastHeartbeat    time.Time

	UploadCount int64
	UploadSum   int64
}

type Registry struct {
	mu      sync.RWMutex
	Devices map[string]*DeviceStats
}

func (r *Registry) AddHeartbeat(deviceID string, sentAt time.Time) error {
	r.mu.Lock()
	stats, exists := r.Devices[deviceID]
	r.mu.Unlock()

	if !exists {
		return fmt.Errorf("Device not found")
	}

	stats.mu.Lock()
	defer stats.mu.Unlock()

	// Assumption: heartbeats dont arrive out of order
	if stats.FirstHeartbeat.IsZero() {
		stats.FirstHeartbeat = sentAt
	}
	if stats.LastHeartbeat.IsZero() || sentAt.After(stats.LastHeartbeat) {
		stats.LastHeartbeat = sentAt
	}

	minute := sentAt.Unix() / 60
	stats.HeartbeatMinutes[minute] = struct{}{}
	return nil
}

func (r *Registry) GetStats(deviceID string) (*DeviceStats, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ds, exists := r.Devices[deviceID]
	return ds, exists
}

func (r *Registry) AddDevice(deviceID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.Devices[deviceID]; !exists {
		r.Devices[deviceID] = &DeviceStats{
			HeartbeatMinutes: make(map[int64]struct{}),
		}
	}
}
