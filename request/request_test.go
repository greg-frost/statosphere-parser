package request

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"statosphere/parser/mock"
)

func TestBarrier(t *testing.T) {
	tests := []struct {
		test     string
		value    int
		pause    time.Duration
		requests uint
		duration time.Duration
		result   bool
	}{
		{"Pass", 5, 0, 10, time.Second, true},
		{"Fail", 15, 0, 10, time.Second, false},
		{"PassPause", 25, 10 * time.Millisecond, 10, 100 * time.Millisecond, true},
		{"FailPause", 25, 1 * time.Millisecond, 10, 100 * time.Millisecond, false},
		{"NoRequests", 5, 0, 0, time.Second, false},
		{"NoTime", 5, 0, 10, 0, true},
	}

	for k, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			var result bool
			for i := 0; i < tt.value; i++ {
				result = Barrier("test_"+fmt.Sprint(k), tt.requests, tt.duration)
				time.Sleep(tt.pause)
			}

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestParamInt(t *testing.T) {
	tests := []struct {
		test   string
		key    string
		get    interface{}
		post   interface{}
		result int
	}{
		{"Get", "get_param", 100, nil, 100},
		{"Copy", "get_param", 150, nil, 150},
		{"Post", "postParam", 100, 200, 200},
		{"Wrong", "wrong.param", "Ten", "", 0},
		{"PartlyWrong", "pwp", "Ten", 100, 100},
		{"Negative", "neg", -300, nil, -300},
		{"Empty", "", 100, 50, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			request := mock.NewRequest()

			if tt.key != "" && tt.get != nil {
				request.AddParamGET(tt.key, tt.get)
			}
			if tt.key != "" && tt.post != nil {
				request.AddParamPOST(tt.key, tt.post)
			}

			result := ParamInt(&request.Request, tt.key, 0)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestParamPositiveInt(t *testing.T) {
	tests := []struct {
		test   string
		key    string
		get    interface{}
		post   interface{}
		result uint
	}{
		{"Get", "get_param", 100, nil, 100},
		{"Wrong", "wrong.param", "Ten", "", 0},
		{"Negative", "neg", -300, nil, 0},
		{"Empty", "", 100, 50, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			request := mock.NewRequest()

			if tt.key != "" && tt.get != nil {
				request.AddParamGET(tt.key, tt.get)
			}
			if tt.key != "" && tt.post != nil {
				request.AddParamPOST(tt.key, tt.post)
			}

			result := ParamPositiveInt(&request.Request, tt.key, 0)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestParamBool(t *testing.T) {
	tests := []struct {
		test   string
		key    string
		get    interface{}
		post   interface{}
		result bool
	}{
		{"True", "TRUE", true, nil, true},
		{"On", "on", nil, "on", true},
		{"Enabled", "enbl", "Enabled", false, false},
		{"False", "FALSE", false, nil, false},
		{"False", "0", nil, "0", false},
		{"NoVal", "", "", "", false},
		{"Empty", "", true, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			request := mock.NewRequest()

			if tt.key != "" && tt.get != nil {
				request.AddParamGET(tt.key, tt.get)
			}
			if tt.key != "" && tt.post != nil {
				request.AddParamPOST(tt.key, tt.post)
			}

			result := ParamBool(&request.Request, tt.key, false)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestParamString(t *testing.T) {
	tests := []struct {
		test   string
		key    string
		get    interface{}
		post   interface{}
		result string
	}{
		{"Get", "GET", "Test", nil, "Test"},
		{"Post", "POST", nil, "Another test", "Another test"},
		{"Copy", "POST", nil, "Copy test", "Copy test"},
		{"Both", "BOTH", "Test 1", "Test 2", "Test 2"},
		{"Number", "NUM", 100, nil, "100"},
		{"Empty", "", "Empty test", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			request := mock.NewRequest()

			if tt.key != "" && tt.get != nil {
				request.AddParamGET(tt.key, tt.get)
			}
			if tt.key != "" && tt.post != nil {
				request.AddParamPOST(tt.key, tt.post)
			}

			result := ParamString(&request.Request, tt.key, "")

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestParamList(t *testing.T) {
	type list map[string][]interface{}

	tests := []struct {
		test   string
		key    string
		get    list
		post   list
		result []string
	}{
		{"Get", "get", list{"get[]": {"Test", "String", 100}}, nil, []string{"Test", "String", "100"}},
		{"Post", "post", list{"get[]": {"Test"}}, list{"post[]": {"String"}}, []string{"String"}},
		{"Both", "param", list{"param[]": {"Test"}}, list{"param[]": {"String"}}, []string{"String", "Test"}},
		{"Wrong", "get", list{"get": {"Test", "String"}}, nil, []string{"Test"}},
		{"List", "list", list{"list": {"[Test, String, 100]"}}, nil, []string{"Test", "String", "100"}},
		{"Empty", "", list{"key": {"value"}}, nil, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			request := mock.NewRequest()

			for k, vv := range tt.get {
				for _, v := range vv {
					if k != "" && v != nil {
						request.AddParamGET(k, v)
					}
				}
			}
			for k, vv := range tt.post {
				for _, v := range vv {
					if k != "" && v != nil {
						request.AddParamPOST(k, v)
					}
				}
			}

			result := ParamList(&request.Request, tt.key, []string{})

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получены значения: %v, ожидаются: %v", result, tt.result)
			}
		})
	}
}
