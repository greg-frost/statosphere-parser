package value

import (
	"reflect"
	"statosphere/parser/mock"
	"strings"
	"testing"
)

var (
	exact  = Value{Exact: 1234}
	approx = Value{Approx: 1200, Short: "1.2K"}
	empty  = Value{}
)

func TestNew(t *testing.T) {
	tests := []struct {
		test    string
		value   string
		isExact bool
		result  Value
		isError bool
	}{
		{"Small", "100", false, Value{Exact: 100}, false},
		{"SmallExact", "200", true, Value{Exact: 200}, false},
		{"Mid", " 12 000 ", false, Value{Approx: 12000, Short: "12 000"}, false},
		{"MidExact", " 12 345 ", true, Value{Exact: 12345}, false},
		{"Factor", "1,2K", false, Value{Approx: 1200, Short: "1,2K"}, false},
		{"FactorExact", "0.2M", false, Value{Approx: 200000, Short: "0.2M"}, false},
		{"Negative", "-500", false, Value{Short: "-500"}, false},
		{"NotNumber", "One", false, Value{}, true},
		{"Empty", "", false, Value{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, err := New(tt.value, tt.isExact)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestValue(t *testing.T) {
	tests := []struct {
		test   string
		value  Value
		result uint
	}{
		{"Exact", exact, 1234},
		{"Approx", approx, 1200},
		{"Empty", empty, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := tt.value.Value()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestIsExact(t *testing.T) {
	tests := []struct {
		test   string
		value  Value
		result bool
	}{
		{"Exact", exact, true},
		{"Approx", approx, false},
		{"Empty", empty, false},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := tt.value.IsExact()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		test   string
		value  Value
		result string
	}{
		{"Exact", exact, "1234"},
		{"Approx", approx, "1200 (1.2K)"},
		{"Empty", empty, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := tt.value.String()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	tests := []struct {
		test    string
		value   Value
		caption string
		result  string
	}{
		{"Exact", exact, "Exact:", "Exact: 1234"},
		{"Approx", approx, "Approx:", "Approx: 1200 (1.2K)"},
		{"Empty", empty, "", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := mock.ReadOutput(func() {
				tt.value.Print(tt.caption)
			})

			if !strings.Contains(result, tt.result) {
				t.Errorf("Получено значение: %q, ожидается совпадение: %q", result, tt.result)
			}
		})
	}
}
