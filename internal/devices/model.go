package devices

import (
	"sync"
	"time"
)

type Device struct {
	ID string `json:"id"`
}

type Heartbeat struct {
	SentAt   time.Time `json:"sent_at"`
	DeviceID string    `json:"device_id"`
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

func (r *Registry) AddHeartbeat(deviceID string, sentAt time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	stats, exists := r.Devices[deviceID]
	if !exists {
		r.Devices[deviceID] = &DeviceStats{
			HeartbeatMinutes: make(map[int64]struct{}),
		}
	}

	stats.mu.Lock()
	defer stats.mu.Unlock()

	stats.LastHeartbeat = sentAt
	if stats.FirstHeartbeat.IsZero() {
		stats.FirstHeartbeat = sentAt
	}
	stats.HeartbeatMinutes[sentAt.Unix()/60] = struct{}{}
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
