package mock

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestNewResponseWriter(t *testing.T) {
	tests := []struct {
		test   string
		result ResponseWriter
	}{
		{"Valid", ResponseWriter{headers: make(http.Header), results: make(map[string]interface{})}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := NewResponseWriter()

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestHeader(t *testing.T) {
	response := ResponseWriter{headers: make(http.Header)}

	tests := []struct {
		test     string
		isChange bool
		result   http.Header
	}{
		{"Valid", false, response.headers},
		{"AlsoValid", true, response.headers},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isChange {
				response.headers.Add("key", "value")
			}
			result := response.Header()

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestWriteHeader(t *testing.T) {
	response := ResponseWriter{results: make(map[string]interface{})}

	tests := []struct {
		test   string
		result int
	}{
		{"Ok", 200},
		{"NotFound", 404},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			response.WriteHeader(tt.result)
			result := response.results["code"]

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	response := ResponseWriter{results: make(map[string]interface{})}

	tests := []struct {
		test    string
		isClear bool
		value   string
		result  string
		count   int
	}{
		{"Valid", false, "Text", "Text", 4},
		{"Append", false, "...", "Text...", 3},
		{"Empty", true, "", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isClear {
				response.ClearBody()
			}

			count, _ := response.Write([]byte(tt.value))
			result := response.results["body"]

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
			if count != tt.count {
				t.Errorf("Получено количество: %v, ожидается: %v", count, tt.count)
			}
		})
	}
}

func TestHead(t *testing.T) {
	response := ResponseWriter{headers: make(http.Header)}

	tests := []struct {
		test   string
		key    string
		result string
	}{
		{"Valid", "Charset", "utf8"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			response.headers.Add(tt.key, tt.result)
			result := response.Head(tt.key)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestCode(t *testing.T) {
	response := ResponseWriter{results: make(map[string]interface{})}

	tests := []struct {
		test   string
		result int
	}{
		{"Ok", 200},
		{"NotFound", 404},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			response.results["code"] = tt.result
			result := response.Code()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestBody(t *testing.T) {
	response := ResponseWriter{results: make(map[string]interface{})}

	tests := []struct {
		test   string
		result string
	}{
		{"Valid", "Text"},
		{"Empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			response.results["body"] = tt.result
			result := response.Body()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestClearBody(t *testing.T) {
	response := ResponseWriter{results: make(map[string]interface{})}

	tests := []struct {
		test    string
		value   string
		isClear bool
		result  string
	}{
		{"Valid", "Text", false, "Text"},
		{"Append", "...", false, "Text..."},
		{"Clear", "***", true, ""},
		{"Empty", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			response.Write([]byte(tt.value))
			if tt.isClear {
				response.ClearBody()
			}
			result := response.Body()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	tests := []struct {
		test   string
		result Request
	}{
		{"Valid", Request{
			Request: http.Request{Method: "GET", URL: &url.URL{}},
			get:     make(url.Values),
			post:    make(url.Values),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := NewRequest()

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestAddParamGET(t *testing.T) {
	request := Request{get: make(url.Values)}

	tests := []struct {
		test   string
		method string
		key    string
		result string
	}{
		{"Valid", "GET", "param", "value"},
		{"Empty", "GET", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			request.AddParamGET(tt.key, tt.result)

			result := request.get[tt.key][0]
			method := request.Method

			if method != tt.method {
				t.Errorf("Method - получено значение: %v, ожидается: %v", method, tt.method)
			}
			if result != tt.result {
				t.Errorf("Value - получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestAddParamPOST(t *testing.T) {
	request := Request{post: make(url.Values)}

	tests := []struct {
		test   string
		method string
		key    string
		result string
	}{
		{"Valid", "POST", "param", "value"},
		{"Empty", "POST", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			request.AddParamPOST(tt.key, tt.result)

			result := request.post[tt.key][0]
			method := request.Method

			if method != tt.method {
				t.Errorf("Method - получено значение: %v, ожидается: %v", method, tt.method)
			}
			if result != tt.result {
				t.Errorf("Value - получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestReadOutput(t *testing.T) {
	tests := []struct {
		test   string
		print  func()
		result string
	}{
		{"Valid", func() { fmt.Println("Test") }, "Test\n"},
		{"Empty", func() {}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := ReadOutput(tt.print)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}
