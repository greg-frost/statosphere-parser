package date

import (
	"fmt"
	"time"
)

// Парсинг даты
func Parse(str string) (time.Time, error) {
	date, err := time.Parse("2006-01-02T15:04:05Z07:00", str)
	if err != nil {
		return time.Time{}, fmt.Errorf("дата %q не прочитана", str)
	}
	return date, nil
}

// Парсинг даты в строку
func ParseString(str string) (string, error) {
	if len(str) < 19 {
		return "", fmt.Errorf("дата %q не прочитана", str)
	}

	return str[:10] + " " + str[11:19], nil
}

// Формирование локальной даты
func Local(date time.Time, hours int) time.Time {
	if hours < -12 || hours > 12 {
		return date
	}

	return date.Add(time.Duration(hours) * time.Hour)
}

// Формирование строковой даты
func String(date time.Time) string {
	return date.Format("2006-01-02 15:04:05")
}
