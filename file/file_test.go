package file

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	var (
		filename = "file_read_temp"
		filetext = "Text"
	)

	Write(filename, filetext)
	defer os.Remove(filename)

	tests := []struct {
		test    string
		file    string
		result  string
		isError bool
	}{
		{"Valid", filename, filetext, false},
		{"NoFile", "no_" + filename, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, err := Read(tt.file)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestReadLines(t *testing.T) {
	var (
		filename = "file_readlines_temp"
		filetext = "Line1\nLine2\nLine3"
	)

	Write(filename, filetext)
	defer os.Remove(filename)

	tests := []struct {
		test    string
		file    string
		result  []string
		isError bool
	}{
		{"Valid", filename, strings.Split(filetext, "\n"), false},
		{"NoFile", "no_" + filename, []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, err := ReadLines(tt.file)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		test    string
		file    string
		value   string
		result  int
		isError bool
	}{
		{"Valid", "file_write_temp", "Test", 4, false},
		{"Invalid", "!@#file.write%/temp^", "Test", 0, true},
		{"NoFile", "", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, err := Write(tt.file, tt.value)
			os.Remove(tt.file)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestChangeDir(t *testing.T) {
	base, _ := os.Getwd()
	temp := filepath.Join(base, "temp")
	inner := filepath.Join(temp, "inner")

	os.MkdirAll(inner, 0755)
	defer os.RemoveAll(temp)

	tests := []struct {
		test   string
		dir    string
		result string
	}{
		{"Valid", "temp", temp},
		{"Slashes", "/temp/", temp},
		{"NoDir", "no_temp", base},
		{"Inner", "inner", base},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			os.Chdir(base)
			ChangeDir(tt.dir)
			result, _ := os.Getwd()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}
