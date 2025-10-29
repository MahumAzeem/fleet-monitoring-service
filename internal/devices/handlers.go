package devices

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

func writeJSONError(w http.ResponseWriter, msg string, statusCode int) {
	WriteJSON(w, ErrorResponse{Msg: msg}, statusCode)
}

func WriteJSON(w http.ResponseWriter, v any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}

func HeartbeatHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, registry *Registry) {
	deviceID := ps.ByName("device_id")

	_, exists := registry.GetStats(deviceID)
	if !exists {
		writeJSONError(w, "Device not found", http.StatusNotFound)
		return
	}

	var hb struct {
		SentAt time.Time `json:"sent_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&hb); err != nil {
		writeJSONError(w, "Invalid json", http.StatusBadRequest)
		return
	}

	if err := registry.AddHeartbeat(deviceID, hb.SentAt); err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func PostStatsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, registry *Registry) {
	deviceID := ps.ByName("device_id")

	var stats StatsRequest
	if err := json.NewDecoder(r.Body).Decode(&stats); err != nil {
		writeJSONError(w, "Invalid json", http.StatusBadRequest)
		return
	}

	ds, exists := registry.GetStats(deviceID)
	if !exists {
		writeJSONError(w, "Device not found", http.StatusNotFound)
		return
	}

	ds.mu.Lock()
	ds.UploadCount++
	ds.UploadSum += stats.UploadTime //in nanoseconds
	ds.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func GetStatsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, registry *Registry) {
	deviceID := ps.ByName("device_id")

	ds, exists := registry.GetStats(deviceID)
	if !exists {
		writeJSONError(w, "Device not found", http.StatusNotFound)
		return
	}

	resp := StatsResponse{
		Uptime:        ds.CalculateUptime(),
		AvgUploadTime: ds.CalculateAvgUploadTime(),
	}
	WriteJSON(w, resp, http.StatusOK)
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	WriteJSON(w, "OK.", http.StatusOK)
}
