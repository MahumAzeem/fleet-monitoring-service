package devices

import (
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
