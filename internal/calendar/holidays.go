package calendar

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const dateLayout = "2006-01-02"

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
