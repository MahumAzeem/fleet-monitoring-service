package devices

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

func HeartbeatHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, registry *Registry) {
	deviceID := ps.ByName("device_id")
	// check if device exists

	_d, exists := registry.GetStats(deviceID)
	fmt.Println("_____________________________________")
	fmt.Printf("Loaded devices: %v\n", registry.Devices)
	fmt.Printf("Loaded device: %v\n", _d)

	if !exists {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}
	var hb struct {
		SentAt time.Time `json:"sent_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&hb); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	registry.AddHeartbeat(deviceID, hb.SentAt)

	w.WriteHeader(http.StatusNoContent)
}

func PostStatsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, registry *Registry) {
	deviceID := ps.ByName("device_id")

	var stats StatsRequest

	if err := json.NewDecoder(r.Body).Decode(&stats); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	ds, exists := registry.GetStats(deviceID)
	if !exists {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	ds.mu.Lock()
	ds.UploadCount++
	ds.UploadSum += int64(stats.UploadTime)
	ds.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func GetStatsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, registry *Registry) {
	deviceID := ps.ByName("device_id")

	ds, exists := registry.GetStats(deviceID)
	if !exists {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var uptime float64
	if ds.FirstHeartbeat.IsZero() || ds.LastHeartbeat.IsZero() {
		uptime = 0.0
	} else if ds.LastHeartbeat.Before(ds.FirstHeartbeat) {
		http.Error(w, "Invalid heartbeat timestamps", http.StatusInternalServerError)
		return
	}
	totalMinutes := int(ds.LastHeartbeat.Sub(ds.FirstHeartbeat).Minutes()) + 1
	uptime = float64(len(ds.HeartbeatMinutes)) / float64(totalMinutes) * 100

	var avgUpload string
	if ds.UploadCount > 0 {
		duration := time.Duration(ds.UploadSum / ds.UploadCount)
		avgUpload = duration.String()
	} else {
		avgUpload = "0"
	}

	resp := StatsResponse{
		Uptime:        uptime,
		AvgUploadTime: avgUpload,
	}

	json.NewEncoder(w).Encode(resp)
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Ok!\n")
}
