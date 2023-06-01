package proxy

import (
	"fmt"
	"os"
	"testing"
	"time"
)

const (
	mainTreshold  uint = 10
	proxyTreshold uint = 3
	cooldown           = 50 * time.Millisecond
)

var (
	validProxies = []string{
		"http://164.92.180.67:8080",
		"https://105.112.191.250:3128",
		"186.232.119.58:3128",
	}
	partlyValidProxies = []string{
		"http://164.92.180.67:8080",
		"Not a proxy, but a text",
		" 186.232.119.58:3128 ",
	}
	invalidProxies = []string{
		"Not a proxy", "  ", "",
	}
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
				Enable(mainTreshold, proxyTreshold, cooldown)
			}

			result := p.isEnabled

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestDisable(t *testing.T) {
	Enable(mainTreshold, proxyTreshold, cooldown)

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

			result := p.isEnabled

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPrepare(t *testing.T) {
	Enable(mainTreshold, proxyTreshold, cooldown)

	tests := []struct {
		test   string
		values []string
		result int
	}{
		{"Valid", validProxies, 3},
		{"PartlyValid", partlyValidProxies, 2},
		{"Invalid", invalidProxies, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			Clear()

			Prepare(tt.values)
			result := len(p.proxies) - 1

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPrepareFromList(t *testing.T) {
	Enable(mainTreshold, proxyTreshold, cooldown)

	tests := []struct {
		test   string
		values []string
		result int
	}{
		{"Valid", validProxies, 3},
		{"PartlyValid", partlyValidProxies, 2},
		{"Invalid", invalidProxies, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			Clear()

			PrepareFromList(tt.values)
			result := len(p.proxies) - 1

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPrepareFromFile(t *testing.T) {
	Enable(mainTreshold, proxyTreshold, cooldown)

	tests := []struct {
		test   string
		values []string
		result int
	}{
		{"Valid", validProxies, 3},
		{"PartlyValid", partlyValidProxies, 2},
		{"Invalid", invalidProxies, 0},
	}

	for i, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			Clear()

			filename := "proxy_temp_" + fmt.Sprint(i)
			file, _ := os.Create(filename)
			for _, proxy := range tt.values {
				file.WriteString(proxy + "\n\n")
			}
			defer os.Remove(filename)
			defer file.Close()

			PrepareFromFile(filename)
			result := len(p.proxies) - 1

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestReset(t *testing.T) {
	Enable(mainTreshold, proxyTreshold, cooldown)

	Clear()
	PrepareFromList(validProxies)

	for i := 0; i < 12; i++ {
		Round()
	}

	tests := []struct {
		test    string
		isReset bool
		result  string
		count   int
	}{
		{"Bypass", false, "http://164.92.180.67:8080", 3},
		{"Reset", true, "", 3},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isReset {
				Reset()
			}

			result := Current()
			count := len(p.proxies) - 1

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
			if count != tt.count {
				t.Errorf("Получено количество: %v, ожидается: %v", count, tt.count)
			}
		})
	}
}

func TestClear(t *testing.T) {
	Enable(mainTreshold, proxyTreshold, cooldown)

	Clear()
	PrepareFromList(validProxies)

	for i := 0; i < 15; i++ {
		Round()
	}

	tests := []struct {
		test    string
		isClear bool
		result  string
		count   int
	}{
		{"Bypass", false, "https://105.112.191.250:3128", 3},
		{"Clear", true, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isClear {
				Clear()
			}

			result := Current()
			count := len(p.proxies) - 1

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
			if count != tt.count {
				t.Errorf("Получено количество: %v, ожидается: %v", count, tt.count)
			}
		})
	}
}

func TestCurrent(t *testing.T) {
	Enable(mainTreshold, proxyTreshold, cooldown)

	Clear()
	PrepareFromList(validProxies)

	tests := []struct {
		test      string
		times     int
		isDisable bool
		result    string
	}{
		{"1 time", 1, false, ""},
		{"10 times", 10, false, ""},
		{"100 times", 100, false, ""},
		{"Disabled", 1, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isDisable {
				Disable()
			}

			Reset()

			var result string
			for i := 0; i < tt.times; i++ {
				result = Current()
			}

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestNext(t *testing.T) {
	Enable(mainTreshold, proxyTreshold, cooldown)

	Clear()
	PrepareFromList(validProxies)

	tests := []struct {
		test      string
		times     int
		isDisable bool
		result    string
	}{
		{"1 time", 1, false, ""},
		{"10 times", 10, false, "http://164.92.180.67:8080"},
		{"20 times", 20, false, "http://186.232.119.58:3128"},
		{"Disabled", 1, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isDisable {
				Disable()
			}

			Reset()

			var result string
			for i := 0; i < tt.times; i++ {
				result = Next()
			}

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestRound(t *testing.T) {
	Enable(mainTreshold, proxyTreshold, cooldown)

	Clear()
	PrepareFromList(validProxies)

	tests := []struct {
		test      string
		times     int
		isDisable bool
		result    string
	}{
		{"1 time", 1, false, ""},
		{"15 times", 15, false, "https://105.112.191.250:3128"},
		{"30 times", 30, false, "http://164.92.180.67:8080"},
		{"Disabled", 1, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isDisable {
				Disable()
			}

			Reset()

			var result string
			for i := 0; i < tt.times; i++ {
				result = Round()
			}

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestCooldown(t *testing.T) {
	Enable(mainTreshold, proxyTreshold, cooldown)

	Clear()
	PrepareFromList(validProxies)

	tests := []struct {
		test      string
		times     int
		waitMs    int
		isDisable bool
		result    string
	}{
		{"1 time", 1, 0, false, ""},
		{"30 times", 30, 0, false, "http://164.92.180.67:8080"},
		{"50 times", 50, 5, false, ""},
		{"20 times", 20, 0, false, "https://105.112.191.250:3128"},
		{"Disabled", 1, 0, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isDisable {
				Disable()
			}

			var result string
			for i := 0; i < tt.times; i++ {
				result = Cooldown()
				time.Sleep(time.Duration(tt.waitMs) * time.Millisecond)
			}

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}
