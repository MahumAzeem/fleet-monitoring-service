package devices

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
)

// func LoadDevicesCSV(path string, registry *Registry) error {
// 	f, err := os.Open(path)
// 	if err != nil {
// 		return fmt.Errorf("failed to open CSV: %w", err)
// 	}
// 	defer f.Close()

// 	reader := csv.NewReader(f)
// 	reader.FieldsPerRecord = -1 // allow variable columns

// 	first := true
// 	count := 0

// 	for {
// 		record, err := reader.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			return fmt.Errorf("error reading CSV: %w", err)
// 		}
// 		if len(record) == 0 {
// 			continue
// 		}

// 		deviceID := strings.TrimSpace(strings.TrimPrefix(record[0], "\uFEFF")) // handle BOM + spaces
// 		if deviceID == "" {
// 			continue
// 		}

// 		// skip header automatically
// 		if first && strings.EqualFold(deviceID, "device_id") {
// 			first = false
// 			continue
// 		}
// 		first = false

// 		registry.AddDevice(deviceID)
// 		count++
// 	}

// 	fmt.Printf("Loaded %d devices from %s\n", count, path)
// 	return nil
// }

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

		deviceID := (record[0]) // handle BOM + spaces
		fmt.Println("DeviceID:", deviceID)
		if deviceID == "" {
			continue
		}

		registry.AddDevice(deviceID)
	}

	return nil
}
