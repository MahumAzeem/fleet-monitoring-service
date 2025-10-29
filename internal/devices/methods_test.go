package devices

import (
	"testing"
	"time"
)

func TestAddDeviceAndGetStats(t *testing.T) {
	reg := NewRegistry()

	reg.AddDevice("test_device")

	ds, exists := reg.GetStats("test_device")
	if !exists {
		t.Fatal("expected device to exist")
	}

	if ds == nil {
		t.Fatal("expected DeviceStats to exist")
	}

	_, exists = reg.GetStats("unknown_device")
	if exists {
		t.Fatal("expected unknown_device to not exist")
	}
}

func TestAddHeartbeat(t *testing.T) {
	now := time.Now()
	reg := NewRegistry()
	reg.AddDevice("test_device")

	err := reg.AddHeartbeat("test_device", now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ds, _ := reg.GetStats("test_device")
	if ds.FirstHeartbeat != now || ds.LastHeartbeat != now {
		t.Fatal("Heartbeat timestamps not set correctly")
	}

	minute := now.Unix() / 60
	if _, ok := ds.HeartbeatMinutes[minute]; !ok {
		t.Fatal("Heartbeat not recorded")
	}

	// Adding heartbeat to non-existent device
	err = reg.AddHeartbeat("unknown_device", now)
	if err == nil {
		t.Fatal("Expected error for unknown device")
	}
}

func TestCalculateUptime(t *testing.T) {
	reg := NewRegistry()
	reg.AddDevice("test_device")
	ds, _ := reg.GetStats("test_device")

	if got := ds.CalculateUptime(); got != 0 {
		t.Fatalf("Expected 0 uptime, got %f", got)
	}

	start := time.Now()
	for i := 1; i < 4; i++ {
		_ = reg.AddHeartbeat("test_device", start.Add(time.Duration(i)*time.Minute))
	}

	uptime := ds.CalculateUptime()
	if uptime <= 0 || uptime > 100 {
		t.Fatalf("expected uptime between 0-100, got %f", uptime)
	}
}

func TestCalculateAvgUploadTime(t *testing.T) {
	reg := NewRegistry()
	reg.AddDevice("test_device")
	ds, _ := reg.GetStats("test_device")

	// No uploads yet
	if got := ds.CalculateAvgUploadTime(); got != "0s" {
		t.Fatalf("Expected 0s, got %s", got)
	}

	// Add some upload times
	ds.mu.Lock()
	ds.UploadCount = 2
	ds.UploadSum = int64(3*time.Minute+30*time.Second) + int64(2*time.Minute)
	ds.mu.Unlock()

	avg := ds.CalculateAvgUploadTime()
	if avg == "0s" {
		t.Fatal("Expected non-zero average upload time")
	}
}
