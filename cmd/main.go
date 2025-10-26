package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"safelyyou-fleet/internal/devices"
)

func main() {

	// Read devices.csv
	// for each device_id create DeviceStats{heartbeatMinutes: map[int64]struct{}}.

	registry := &devices.Registry{
		Devices: make(map[string]*devices.DeviceStats),
	}
	fmt.Println("Loading devices from CSV...")

	if err := devices.LoadDevicesCSV("devices.csv", registry); err != nil {
		panic(err)
	}

	baseUrl := "/api/v1"
	router := httprouter.New()
	router.GET("/", devices.Index)

	// POST /devices/{device_id}/stats -- sent_at and upload_time
	router.POST(baseUrl+"/devices/:device_id/stats", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		devices.PostStatsHandler(w, r, ps, registry)
	})

	// GET /devices/{device_id}/stats -- return uptime, avg_upload_time
	router.GET(baseUrl+"/devices/:device_id/stats", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		devices.GetStatsHandler(w, r, ps, registry)
	})

	// POST /devices/{device_id}/heartbeat
	router.POST(baseUrl+"/devices/:device_id/heartbeat", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		devices.HeartbeatHandler(w, r, ps, registry)
	})

	http.ListenAndServe(":8080", router)
}

func wrapHandler(
	h func(http.ResponseWriter, *http.Request, httprouter.Params, *devices.Registry),
	reg *devices.Registry,
) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h(w, r, ps, reg)
	}
}

func wrapHandler(h func(http.ResponseWriter, *http.Request, httprouter.Params, *devices.Registry), reg *devices.Registry) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h(w, r, ps, reg)
	}
}
