package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"safelyyou-fleet/internal/devices"
)

func main() {

	registry := &devices.Registry{
		Devices: make(map[string]*devices.DeviceStats),
	}
	log.Println("Loading devices from CSV...")

	if err := devices.LoadDevicesCSV("devices.csv", registry); err != nil {
		log.Fatalf("Failed to load devices from CSV: %v", err)
	}

	baseURL := "/api/v1"
	router := httprouter.New()

	router.GET("/", devices.Index)
	router.POST(baseURL+"/devices/:device_id/stats", deviceHandler(devices.PostStatsHandler, registry))
	router.GET(baseURL+"/devices/:device_id/stats", deviceHandler(devices.GetStatsHandler, registry))
	router.POST(baseURL+"/devices/:device_id/heartbeat", deviceHandler(devices.HeartbeatHandler, registry))

	log.Println("Server starting on :6733")
	if err := http.ListenAndServe(":6733", router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func deviceHandler(
	h func(http.ResponseWriter, *http.Request, httprouter.Params, *devices.Registry),
	reg *devices.Registry) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h(w, r, ps, reg)
	}
}
