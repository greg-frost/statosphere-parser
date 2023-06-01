package get

import (
	"os"
	"strings"
	"testing"
	"time"
)

func init() {
	os.Chdir("..")
}

func TestPage(t *testing.T) {
	IsCacheDisable = true

	tests := []struct {
		test      string
		transport string
		value     string
		code      int
		result    string
		isError   bool
	}{
		{"WrongHttp", "Http", "not an address", 500, "", true},
		{"EmptyCurl", "Curl", "", 500, "", true},
		{"EmptyFile", "File", "", 500, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			SetTransport(tt.transport, 0)

			code, result, err := Page(tt.value)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if code != tt.code {
				t.Errorf("Получен код: %v, ожидается: %v", code, tt.code)
			}
			if !strings.Contains(result, tt.result) {
				t.Errorf("Получено значение: %q, ожидается совпадение: %q", result, tt.result)
			}
		})
	}
}

func TestPageHTTP(t *testing.T) {
	SetTransport("http", 0)
	IsCacheDisable = true

	tests := []struct {
		test      string
		value     string
		proxy     string
		timeoutMs int
		code      int
		result    string
		isError   bool
	}{
		{"Valid", "https://t.me/thecodemedia", "", 0, 200, "thecodemedia", false},
		{"ProxyTimeout", "http://t.me/thecodemedia", "http://164.92.180.67:8080", 50, 500, "", true},
		{"Wrong", "not an address", "not a proxy", 0, 500, "", true},
		{"Empty", "", "", 0, 500, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			SetTimeout(time.Duration(tt.timeoutMs) * time.Millisecond)

			code, result, err := PageHTTP(tt.value, tt.proxy)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if code != tt.code {
				t.Errorf("Получен код: %v, ожидается: %v", code, tt.code)
			}
			if !strings.Contains(result, tt.result) {
				t.Errorf("Получено значение: %q, ожидается совпадение: %q", result, tt.result)
			}
		})
	}
}

func TestPageCURL(t *testing.T) {
	SetTransport("curl", 0)
	IsCacheDisable = true

	tests := []struct {
		test      string
		value     string
		proxy     string
		timeoutMs int
		code      int
		result    string
		isError   bool
	}{
		{"Valid", "https://t.me/thecodemedia", "", 0, 200, "thecodemedia", false},
		{"ProxyTimeout", "http://t.me/thecodemedia", "http://164.92.180.67:8080", 50, 500, "", true},
		{"NotAddress", "not an address", "", 0, 500, "", true},
		{"NotProxy", "http://t.me/thecodemedia", "not a proxy", 0, 500, "", true},
		{"Empty", "", "", 0, 500, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			SetTimeout(time.Duration(tt.timeoutMs) * time.Millisecond)

			code, result, err := PageCURL(tt.value, tt.proxy)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if code != tt.code {
				t.Errorf("Получен код: %v, ожидается: %v", code, tt.code)
			}
			if !strings.Contains(result, tt.result) {
				t.Errorf("Получено значение: %q, ожидается совпадение: %q", result, tt.result)
			}
		})
	}
}

func TestPageFile(t *testing.T) {
	SetTransport("file", 0)

	tests := []struct {
		test    string
		value   string
		remove  string
		result  int
		isError bool
	}{
		{"Username", "https://t.me/codecamp", "", 200, false},
		{"Joinchat", "https://t.me/+so8YUpEsL4BkZGQy", "", 200, false},
		{"Remove", "https://t.me/serious_tester", "data/pages/info/serious_tester", 200, false},
		{"Messages", "https://t.me/s/thecodemedia", "", 200, false},
		{"Wrong", "***", "", 500, true},
		{"Empty", "", "", 500, true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.remove != "" {
				os.Remove(tt.remove)
			}

			result, _, err := PageFile(tt.value, "")

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if result != tt.result {
				t.Errorf("Получен результат: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestSetTransport(t *testing.T) {
	CurlFasterTreshold = 10

	tests := []struct {
		test   string
		value  string
		count  int
		result string
	}{
		{"Http", "Http", 0, "http"},
		{"HttpHigh", "http", 1000, "http"},
		{"CurlFaster", "Curl", 5, "curl"},
		{"CurlSlower", "CURL", 15, "http"},
		{"File", "file", 0, "file"},
		{"FileHigh", "FiLe", 500, "file"},
		{"Wrong", "tcp", 0, "http"},
		{"Empty", "", 0, "http"},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			SetTransport(tt.value, tt.count)
			result := o.transport

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestTransport(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Http", "Http", "http"},
		{"Curl", "CURL", "curl"},
		{"File", "file", "file"},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			SetTransport(tt.value, 0)
			result := Transport()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestSetTimeout(t *testing.T) {
	tests := []struct {
		test   string
		result time.Duration
	}{
		{"Zero", 0},
		{"Millisecond", time.Millisecond},
		{"Second", time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			SetTimeout(tt.result)
			result := o.timeout

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestTimeout(t *testing.T) {
	tests := []struct {
		test   string
		result time.Duration
	}{
		{"Zero", 0},
		{"Millisecond", time.Millisecond},
		{"Second", time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			o.timeout = tt.result
			result := Timeout()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestTimeoutString(t *testing.T) {
	tests := []struct {
		test   string
		value  time.Duration
		result string
	}{
		{"Zero", 0, "0"},
		{"Millisecond", time.Millisecond, "0"},
		{"Milliseconds", 25 * time.Millisecond, "0,03"},
		{"Second", time.Second, "1"},
		{"Seconds", 1500 * time.Millisecond, "1,50"},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			o.timeout = tt.value
			result := TimeoutString()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}
