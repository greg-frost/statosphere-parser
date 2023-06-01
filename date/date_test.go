package date

import (
	"testing"
	"time"
)

var (
	dateShort        = "06-01-02 15:04"
	dateMid          = "2006-01-02 15:04:05"
	dateMidLocal     = "2006-01-02 18:04:05"
	dateFull         = "2006-01-02T15:04:05+00:00"
	dateFullLocal    = "2006-01-02T18:04:05+00:00"
	dateTime, _      = time.Parse("2006-01-02T15:04:05Z07:00", dateFull)
	dateTimeLocal, _ = time.Parse("2006-01-02T15:04:05Z07:00", dateFullLocal)
)

func TestParse(t *testing.T) {
	tests := []struct {
		test    string
		date    string
		result  time.Time
		isError bool
	}{
		{"Valid", dateFull, dateTime, false},
		{"Wrong", dateMid, time.Time{}, true},
		{"Text", "not a date", time.Time{}, true},
		{"Empty", "", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, err := Parse(tt.date)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		test    string
		date    string
		result  string
		isError bool
	}{
		{"ValidFull", dateFull, dateMid, false},
		{"ValidShort", dateMid, dateMid, false},
		{"Wrong", dateShort, "", true},
		{"Text", "not a date", "", true},
		{"Empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, err := ParseString(tt.date)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestLocal(t *testing.T) {
	tests := []struct {
		test   string
		date   time.Time
		hours  int
		result time.Time
	}{
		{"UTC+0", dateTime, 0, dateTime},
		{"UTC+3", dateTime, 3, dateTimeLocal},
		{"UTC+13", dateTime, 13, dateTime},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Local(tt.date, tt.hours)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		test   string
		date   time.Time
		result string
	}{
		{"Valid", dateTime, dateMid},
		{"ValidLocal", dateTimeLocal, dateMidLocal},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := String(tt.date)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}
