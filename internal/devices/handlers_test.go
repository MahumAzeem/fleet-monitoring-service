package devices

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

const (
	testDeviceID    = "test_device"
	unknownDeviceID = "unknown_device"
)

func createBody(t *testing.T, data interface{}) *bytes.Reader {
	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	return bytes.NewReader(jsonData)
}

func createRequest(method, deviceID string, body *bytes.Reader) (*http.Request, httprouter.Params) {
	path := "/devices/" + deviceID
	var reqBody io.Reader
	if body != nil {
		reqBody = body
	}
	req := httptest.NewRequest(method, path, reqBody)
	params := httprouter.Params{
		httprouter.Param{Key: "device_id", Value: deviceID},
	}
	return req, params
}

func setupRegistryWithDevice() *Registry {
	reg := NewRegistry()
	reg.AddDevice(testDeviceID)
	return reg
}

func getStatsResponse(t *testing.T, reg *Registry, deviceID string) (int, *StatsResponse) {
	req, params := createRequest(http.MethodGet, deviceID, nil)
	recorder := httptest.NewRecorder()
	GetStatsHandler(recorder, req, params, reg)

	var stats StatsResponse
	if recorder.Code == http.StatusOK {
		err := json.NewDecoder(recorder.Body).Decode(&stats)
		assert.NoError(t, err)
		return recorder.Code, &stats
	}
	return recorder.Code, nil
}
func TestHeartbeatHandler(t *testing.T) {
	reg := setupRegistryWithDevice()

	heartbeatPayload := map[string]string{
		"sent_at": time.Now().Format(time.RFC3339),
	}

	testCases := []struct {
		name           string
		deviceID       string
		expectedStatus int
	}{
		{"Existing device", testDeviceID, http.StatusNoContent},
		{"Unknown device", unknownDeviceID, http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := createBody(t, heartbeatPayload)
			req, params := createRequest(http.MethodPost, tc.deviceID, requestBody)
			recorder := httptest.NewRecorder()

			HeartbeatHandler(recorder, req, params, reg)

			assert.Equal(t, tc.expectedStatus, recorder.Code)
		})
	}
}

func TestPostStatsHandler(t *testing.T) {
	reg := setupRegistryWithDevice()

	uploadTime := int64(123456)
	statsPayload := StatsRequest{
		SentAt:     time.Now(),
		UploadTime: uploadTime,
	}

	requestBody := createBody(t, statsPayload)
	req, params := createRequest(http.MethodPost, testDeviceID, requestBody)
	recorder := httptest.NewRecorder()

	PostStatsHandler(recorder, req, params, reg)

	assert.Equal(t, http.StatusNoContent, recorder.Code)

	deviceStats, exists := reg.GetStats(testDeviceID)
	assert.True(t, exists)
	assert.Equal(t, int64(1), deviceStats.UploadCount)
	assert.Equal(t, uploadTime, deviceStats.UploadSum)
}

func TestGetStatsHandler(t *testing.T) {
	t.Run("Unknown device", func(t *testing.T) {
		reg := NewRegistry()
		statusCode, _ := getStatsResponse(t, reg, unknownDeviceID)
		assert.Equal(t, http.StatusNotFound, statusCode)
	})

	t.Run("No heartbeats", func(t *testing.T) {
		reg := setupRegistryWithDevice()
		statusCode, stats := getStatsResponse(t, reg, testDeviceID)

		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, 0.0, stats.Uptime)
		assert.Equal(t, "0s", stats.AvgUploadTime)
	})

	t.Run("First heartbeat == last heartbeat", func(t *testing.T) {
		reg := setupRegistryWithDevice()
		now := time.Now()
		reg.AddHeartbeat(testDeviceID, now)

		statusCode, stats := getStatsResponse(t, reg, testDeviceID)

		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, float64(100), stats.Uptime)
	})

	t.Run("Last heartbeat after first", func(t *testing.T) {
		reg := setupRegistryWithDevice()
		start := time.Now().Add(-3 * time.Minute)
		reg.AddHeartbeat(testDeviceID, start)
		reg.AddHeartbeat(testDeviceID, start.Add(1*time.Minute))
		reg.AddHeartbeat(testDeviceID, start.Add(2*time.Minute))
		reg.AddHeartbeat(testDeviceID, start.Add(3*time.Minute))

		statusCode, stats := getStatsResponse(t, reg, testDeviceID)

		assert.Equal(t, http.StatusOK, statusCode)
		assert.Greater(t, stats.Uptime, 0.0)
	})

	t.Run("2 second average upload time", func(t *testing.T) {
		reg := setupRegistryWithDevice()
		deviceStats, _ := reg.GetStats(testDeviceID)
		deviceStats.UploadCount = 3
		deviceStats.UploadSum = int64(6 * time.Second)

		statusCode, stats := getStatsResponse(t, reg, testDeviceID)

		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, "2s", stats.AvgUploadTime)
	})
}
