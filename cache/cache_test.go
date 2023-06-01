package cache

import (
	"reflect"
	"testing"
	"time"
)

func TestEnable(t *testing.T) {
	Disable()

	tests := []struct {
		test     string
		isEnable bool
		result   bool
	}{
		{"Enable", true, true},
		{"Enabled", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isEnable {
				Enable()
			}

			result := c.isEnabled

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestDisable(t *testing.T) {
	Enable()

	tests := []struct {
		test      string
		isDisable bool
		result    bool
	}{
		{"Disable", true, false},
		{"Disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isDisable {
				Disable()
			}

			result := c.isEnabled

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestSetValue(t *testing.T) {
	Enable()
	Clear()

	tests := []struct {
		test      string
		key       string
		value     interface{}
		expires   time.Duration
		isDisable bool
		result    bool
	}{
		{"String", "str", "value", time.Second, false, true},
		{"Empty", "", nil, time.Millisecond, false, false},
		{"Disabled", "...", "...", time.Nanosecond, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isDisable {
				Disable()
			}

			result := SetValue(tt.key, tt.value, tt.expires)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestValue(t *testing.T) {
	Enable()

	SetValue("str", "value", time.Second)
	SetValue("INT", 1000, 5*time.Second)
	SetValue("Expired", []string{}, 0)

	tests := []struct {
		test      string
		key       string
		result    interface{}
		isExist   bool
		isDisable bool
	}{
		{"String", "str", "value", true, false},
		{"Integer", "INT", 1000, true, false},
		{"Expired", "Expired", nil, false, false},
		{"Empty", "", nil, false, false},
		{"Disabled", "...", nil, false, true},
	}

	time.Sleep(time.Millisecond)

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isDisable {
				Disable()
			}

			result, isExist := Value(tt.key)

			if isExist != tt.isExist {
				t.Fatalf("Получен результат: %v, ожидается: %v", isExist, tt.isExist)
			}
			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v (%T), ожидается: %v (%T)", result, result, tt.result, tt.result)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	Enable()
	Clear()

	SetValue("str", "value", time.Second)
	SetValue("INT", 1000, 5*time.Second)
	SetValue("Expired", []string{}, 0)

	tests := []struct {
		test      string
		key       string
		isDisable bool
		result    int
	}{
		{"NoKey", "no", false, 3},
		{"String", "str", false, 2},
		{"Empty", "", false, 2},
		{"Disabled", "...", true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isDisable {
				Disable()
			}

			Remove(tt.key)
			result := len(c.values)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestClear(t *testing.T) {
	Enable()
	Clear()

	SetValue("Key", "Value", 5*time.Millisecond)

	Value("Key")
	Value("None")

	tests := []struct {
		test    string
		isClear bool
		count   int
		success int
		failed  int
	}{
		{"None", false, 1, 1, 1},
		{"Clear", true, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isClear {
				Clear()
			}

			count, success, failed := len(c.values), c.success, c.failed

			if count != tt.count {
				t.Errorf("Count - получено значение: %v, ожидается: %v", count, tt.count)
			}
			if success != tt.success {
				t.Errorf("Success - получено значение: %v, ожидается: %v", success, tt.success)
			}
			if failed != tt.failed {
				t.Errorf("Failed - получено значение: %v, ожидается: %v", failed, tt.failed)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	Enable()
	Clear()

	SetValue("str", "value", 50*time.Millisecond)
	SetValue("INT", 1000, 100*time.Millisecond)
	SetValue("Expired", []string{}, 0)

	tests := []struct {
		test      string
		waitMs    int
		isDisable bool
		result    int
	}{
		{"First", 10, false, 2},
		{"Second", 50, false, 1},
		{"Third", 15, false, 1},
		{"Disabled", 0, true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isDisable {
				Disable()
			}

			time.Sleep(time.Duration(tt.waitMs) * time.Millisecond)

			Check()
			result := len(c.values)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestCheckEvery(t *testing.T) {
	Enable()
	Clear()

	SetValue("str", "value", 50*time.Millisecond)
	SetValue("INT", 1000, 100*time.Millisecond)
	SetValue("Expired", []string{}, 0)

	interval := 30 * time.Millisecond

	cancel := CheckEvery(interval)

	tests := []struct {
		test     string
		isCancel bool
		result   int
	}{
		{"First", false, 3},
		{"Second", false, 2},
		{"Cancel", true, 1},
		{"Third", false, 1},
		{"Last", false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isCancel {
				cancel()
			}

			result := len(c.values)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}

			time.Sleep(interval)
		})
	}
}

func TestStats(t *testing.T) {
	Enable()
	Clear()

	SetValue("Key", "Value", 5*time.Millisecond)

	Value("Key")
	Value("None")

	tests := []struct {
		test    string
		waitMs  int
		count   int
		success int
		failed  int
	}{
		{"Now", 0, 1, 1, 1},
		{"Wait", 10, 0, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			time.Sleep(time.Duration(tt.waitMs) * time.Millisecond)

			Check()
			count, success, failed := Stats()

			if count != tt.count {
				t.Errorf("Count - получено значение: %v, ожидается: %v", count, tt.count)
			}
			if success != tt.success {
				t.Errorf("Success - получено значение: %v, ожидается: %v", success, tt.success)
			}
			if failed != tt.failed {
				t.Errorf("Failed - получено значение: %v, ожидается: %v", failed, tt.failed)
			}
		})
	}
}
