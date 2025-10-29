package devices

import (
	"fmt"
	"time"
)

func NewRegistry() *Registry {
	return &Registry{
		Devices: make(map[string]*DeviceStats),
	}
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

func (ds *DeviceStats) CalculateUptime() float64 {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if ds.FirstHeartbeat.IsZero() || ds.LastHeartbeat.IsZero() {
		return 0
	}

	totalMinutes := ds.LastHeartbeat.Sub(ds.FirstHeartbeat).Minutes() + 1
	if totalMinutes <= 0 {
		return 0
	}

	return float64(len(ds.HeartbeatMinutes)) / float64(totalMinutes) * 100
}

func (ds *DeviceStats) CalculateAvgUploadTime() string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if ds.UploadCount == 0 {
		return "0s"
	}

	avg := time.Duration(ds.UploadSum / ds.UploadCount)
	return avg.String()
}
