package entities

import "time"

// FrequencyIntervals maps canonical frequency keys to their durations.
// These match the normalized values from migration 000006_normalize_frequencies.
var FrequencyIntervals = map[string]time.Duration{
	"5m":   5 * time.Minute,
	"10m":  10 * time.Minute,
	"15m":  15 * time.Minute,
	"30m":  30 * time.Minute,
	"1h":   1 * time.Hour,
	"2h":   2 * time.Hour,
	"4h":   4 * time.Hour,
	"6h":   6 * time.Hour,
	"12h":  12 * time.Hour,
	"24h":  24 * time.Hour,
	"168h": 168 * time.Hour,
}

// FrequencyAliases maps old verbose format strings to canonical short keys.
var FrequencyAliases = map[string]string{
	"Every 5 minutes":  "5m",
	"Every 10 minutes": "10m",
	"Every 15 minutes": "15m",
	"Every 30 minutes": "30m",
	"Every hour":       "1h",
	"Every 1 hour":     "1h",
	"1 hr":             "1h",
	"Every 2 hours":    "2h",
	"2 hr":             "2h",
	"Every 4 hours":    "4h",
	"4 hr":             "4h",
	"Every 6 hours":    "6h",
	"6 hr":             "6h",
	"Every 12 hours":   "12h",
	"12 hr":            "12h",
	"Every day":        "24h",
	"1d":               "24h",
	"Every week":       "168h",
	"7d":               "168h",
}

// FrequencyToPostgresInterval maps canonical keys to PostgreSQL interval strings.
var FrequencyToPostgresInterval = map[string]string{
	"5m":   "5 minutes",
	"10m":  "10 minutes",
	"15m":  "15 minutes",
	"30m":  "30 minutes",
	"1h":   "1 hour",
	"2h":   "2 hours",
	"4h":   "4 hours",
	"6h":   "6 hours",
	"12h":  "12 hours",
	"24h":  "1 day",
	"168h": "7 days",
}

// ResolveFrequency returns the duration for a given frequency string,
// handling both canonical short keys and verbose aliases.
func ResolveFrequency(freq string) (time.Duration, bool) {
	if d, ok := FrequencyIntervals[freq]; ok {
		return d, true
	}
	if canonical, ok := FrequencyAliases[freq]; ok {
		return FrequencyIntervals[canonical], true
	}
	return 0, false
}

// AllFrequencyKeys returns all keys (canonical + aliases) grouped by canonical key.
func AllFrequencyKeys() map[string][]string {
	allKeys := make(map[string][]string)
	for canonical := range FrequencyIntervals {
		allKeys[canonical] = append(allKeys[canonical], canonical)
	}
	for alias, canonical := range FrequencyAliases {
		allKeys[canonical] = append(allKeys[canonical], alias)
	}
	return allKeys
}
