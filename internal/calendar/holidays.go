package calendar

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"gopkg.in/yaml.v3"
)

const dateLayout = "2006-01-02"
const syukujitsuDateLayout = "2006/1/2"

// LoadHolidaysYAML reads holiday dates from a YAML file and registers them.
// Supported formats:
//   - A top-level list of YYYY-MM-DD strings
//   - A map with key "holidays" pointing to a list of YYYY-MM-DD strings
func LoadHolidaysYAML(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read holidays yaml: %w", err)
	}

	dates, err := parseHolidayDates(data)
	if err != nil {
		return err
	}
	SetHolidays(dates)
	return nil
}

// LoadHolidaysFile reads holidays from a supported file and registers them.
// Supported formats:
//   - YAML with YYYY-MM-DD strings
//   - Cabinet Office syukujitsu.csv
func LoadHolidaysFile(path string) error {
	ext := filepath.Ext(path)
	switch ext {
	case ".csv":
		return LoadHolidaysCSV(path)
	case ".yaml", ".yml":
		return LoadHolidaysYAML(path)
	default:
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read holidays file: %w", err)
		}
		if dates, err := parseHolidayCSV(data); err == nil {
			SetHolidays(dates)
			return nil
		}
		dates, err := parseHolidayDates(data)
		if err != nil {
			return fmt.Errorf("unsupported holidays file format: %w", err)
		}
		SetHolidays(dates)
		return nil
	}
}

// LoadHolidaysCSV reads holiday dates from Cabinet Office syukujitsu.csv and registers them.
func LoadHolidaysCSV(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read holidays csv: %w", err)
	}

	dates, err := parseHolidayCSV(data)
	if err != nil {
		return err
	}
	SetHolidays(dates)
	return nil
}

func parseHolidayDates(data []byte) ([]time.Time, error) {
	var withKey struct {
		Holidays []string `yaml:"holidays"`
	}
	if err := yaml.Unmarshal(data, &withKey); err == nil && len(withKey.Holidays) > 0 {
		return parseDateStrings(withKey.Holidays)
	}

	var list []string
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("decode holidays yaml: %w", err)
	}
	return parseDateStrings(list)
}

func parseHolidayCSV(data []byte) ([]time.Time, error) {
	decoded, err := decodeHolidayCSVBytes(data)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(decoded))
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("decode holidays csv: empty file")
		}
		return nil, fmt.Errorf("decode holidays csv: %w", err)
	}
	if len(header) == 0 {
		return nil, fmt.Errorf("decode holidays csv: missing header")
	}
	if header[0] != "国民の祝日・休日月日" {
		return nil, fmt.Errorf("decode holidays csv: unexpected header %q", header[0])
	}

	var dates []time.Time
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode holidays csv: %w", err)
		}
		if len(record) == 0 || record[0] == "" {
			continue
		}
		parsed, err := time.ParseInLocation(syukujitsuDateLayout, record[0], time.Local)
		if err != nil {
			return nil, fmt.Errorf("parse holiday %q: %w", record[0], err)
		}
		dates = append(dates, parsed)
	}
	return dates, nil
}

func parseDateStrings(values []string) ([]time.Time, error) {
	dates := make([]time.Time, 0, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		parsed, err := time.ParseInLocation(dateLayout, v, time.Local)
		if err != nil {
			return nil, fmt.Errorf("parse holiday %q: %w", v, err)
		}
		dates = append(dates, parsed)
	}
	return dates, nil
}

func decodeHolidayCSVBytes(data []byte) ([]byte, error) {
	trimmed := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if utf8.Valid(trimmed) {
		return trimmed, nil
	}
	utf8Reader := transform.NewReader(bytes.NewReader(data), japanese.ShiftJIS.NewDecoder())
	decoded, err := io.ReadAll(utf8Reader)
	if err == nil {
		return bytes.TrimPrefix(decoded, []byte{0xEF, 0xBB, 0xBF}), nil
	}
	return nil, fmt.Errorf("decode holidays csv: %w", err)
}
