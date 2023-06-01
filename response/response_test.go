package response

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"statosphere/parser/mock"
)

func TestNew(t *testing.T) {
	tests := []struct {
		test   string
		value  interface{}
		count  int
		errs   []error
		time   time.Duration
		result Response
	}{
		{"Valid", "Test", 1, nil, 100 * time.Millisecond,
			Response{Ok: true, Code: 200, Status: "OK", Data: "Test", Time: 100 * time.Millisecond}},
		{"NotFound", nil, 0, []error{errors.New("Not found")}, 0,
			Response{Ok: false, Code: 404, Status: "Not Found", Errors: []error{errors.New("Not found")}}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := New(tt.value, tt.count, tt.errs, tt.time)

			if result.Ok != tt.result.Ok || result.Code != tt.result.Code ||
				!reflect.DeepEqual(result.Data, tt.result.Data) ||
				len(result.Errors) != len(tt.result.Errors) ||
				result.Time != tt.result.Time {
				t.Errorf("Получено значение: %#v, ожидается: %#v", result, tt.result)
			}
		})
	}
}

func TestToStrings(t *testing.T) {
	tests := []struct {
		test      string
		err       error
		errs      []error
		time      time.Duration
		isConvert bool
		result    Response
	}{
		{"Valid", errors.New("Warning"), []error{errors.New("Wrong"), errors.New("Format")}, time.Millisecond,
			true, Response{ErrorMsg: "Warning", ErrorsMsg: []string{"Wrong", "Format"}, TimeMsg: "1ms"}},
		{"None", errors.New("Warning"), []error{errors.New("Wrong"), errors.New("Format")}, time.Millisecond,
			false, Response{ErrorMsg: "", ErrorsMsg: nil, TimeMsg: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Response{Error: tt.err, Errors: tt.errs, Time: tt.time}
			if tt.isConvert {
				result.ToStrings()
			}

			if result.ErrorMsg != tt.result.ErrorMsg ||
				!reflect.DeepEqual(result.ErrorsMsg, tt.result.ErrorsMsg) ||
				result.TimeMsg != tt.result.TimeMsg {
				t.Errorf("Получено значение: %#v, ожидается: %#v", result, tt.result)
			}
		})
	}
}

func TestToValues(t *testing.T) {
	tests := []struct {
		test      string
		err       string
		errs      []string
		time      string
		isConvert bool
		result    Response
	}{
		{"Valid", "Warning", []string{"Wrong format"}, "0s", true,
			Response{Error: errors.New("Warning"), Errors: []error{errors.New("Wrong format")}}},
		{"None", "", nil, "", false, Response{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Response{ErrorMsg: tt.err, ErrorsMsg: tt.errs, TimeMsg: tt.time}
			if tt.isConvert {
				result.ToValues()
			}

			if fmt.Sprint(result.Error) != fmt.Sprint(tt.result.Error) ||
				len(result.Errors) != len(tt.result.Errors) ||
				result.Time != tt.result.Time {
				t.Errorf("Получено значение: %#v, ожидается: %#v", result, tt.result)
			}
		})
	}
}

func TestEncodeJSON(t *testing.T) {
	tests := []struct {
		test    string
		value   Response
		result  string
		isError bool
	}{
		{"Valid", Response{Ok: true, Code: 200, Status: "OK", Data: `Text "value"`, Time: 100 * time.Millisecond,
			Error: errors.New("Warning"), Errors: []error{errors.New("Wrong"), errors.New("Format")}},
			`{"ok":true,"code":200,"status":"OK","data":"Text \"value\"","error":"Warning",` +
				`"errors":["Wrong","Format"],"time":"100ms"}`, false},
		{"Empty", Response{}, `{"ok":false,"code":0,"status":"","time":"0s"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, err := tt.value.EncodeJSON()

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		test    string
		value   string
		result  Response
		isError bool
	}{
		{"Valid", `{"ok":true,"code":200,"status":"OK","data":"Text \"value\"","error":"Warning",` +
			`"errors":["Wrong","Format"],"time":"100ms"}`, Response{Ok: true, Code: 200, Status: "OK",
			Data: `Text "value"`, Time: 100 * time.Millisecond, Error: errors.New("Warning"),
			Errors: []error{errors.New("Wrong"), errors.New("Format")}}, false},
		{"Wrong", "wrong json format", Response{}, true},
		{"Empty", "{}", Response{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Response{}
			err := result.DecodeJSON(tt.value)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if result.Ok != tt.result.Ok || result.Code != tt.result.Code ||
				!reflect.DeepEqual(result.Data, tt.result.Data) ||
				fmt.Sprint(result.Error) != fmt.Sprint(tt.result.Error) ||
				len(result.Errors) != len(tt.result.Errors) ||
				result.Time != tt.result.Time {
				t.Errorf("Получено значение: %#v, ожидается: %#v", result, tt.result)
			}
		})
	}
}

func TestPrintJSON(t *testing.T) {
	tests := []struct {
		test   string
		header string
		result string
	}{
		{"Valid", "application/json", `{"ok":true,"code":200,"status":"OK","data":"Text \"value\""}`},
		{"Empty", "application/json", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			response := mock.NewResponseWriter()

			PrintJSON(response, tt.result)

			header := response.Head("Content-Type")
			result := response.Body()

			if header != tt.header {
				t.Errorf("Получен заголовок: %v, ожидается: %v", header, tt.header)
			}
			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPrintStatus(t *testing.T) {
	tests := []struct {
		test   string
		code   int
		result string
	}{
		{"Ok", 200, "200 OK"},
		{"NotFound", 404, "404 Not Found"},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			response := mock.NewResponseWriter()

			PrintStatus(response, tt.code)

			code := response.Code()
			result := response.Body()

			if code != tt.code {
				t.Errorf("Получен заголовок: %v, ожидается: %v", code, tt.code)
			}
			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}
