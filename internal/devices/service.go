package devices

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"strings"
)

func LoadDevicesCSV(path string, registry *Registry) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true
	r.ReuseRecord = true

	header, err := r.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("CSV file is empty")
		}
		return err
	}
	if len(header) == 0 {
		return errors.New("CSV header is empty")
	}

	for {
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if len(record) == 0 {
			continue
		}

		deviceID := strings.TrimSpace(record[0])
		if deviceID == "" {
			continue
		}

		log.Printf("Registering device ID: %s", deviceID)
		registry.AddDevice(deviceID)
	}

	return nil
}
