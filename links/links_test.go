package links

import (
	"reflect"
	"statosphere/parser/mock"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		test   string
		result Links
	}{
		{"Valid", Links{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := New()

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestKey(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Bypass", "www.example.com", "www.example.com"},
		{"LowerCase", "www.EXAMPLE.com", "www.example.com"},
		{"TrimLink", "https://www.EXAMPLE.com/", "www.example.com"},
		{"QueryString", "https://www.EXAMPLE.com/?param=value", "www.example.com"},
		{"Long", "www.example.com/very/very/verry/verryy/long", "www.example.com/very/very/verry/..."},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Key(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	link := New()
	link.Add("https://www.example.com", "", 0)

	caption := New()
	caption.Add("https://www.example.ru", "Example", 0)

	couple := New()
	couple.Add("https://www.example.com", "", 0)
	couple.Add("https://www.example.com", "", 0)

	post := New()
	post.Add("https://www.site.com", "", 100)

	empty := New()

	tests := []struct {
		test       string
		link       string
		caption    string
		post       uint
		multiplier int
		isAdded    bool
		result     Links
	}{
		{"Link", "https://www.example.com", "", 0, 1, true, link},
		{"Caption", "https://www.example.ru", "Example", 0, 1, true, caption},
		{"ShortCaption", "https://www.example.com", "?", 0, 1, true, link},
		{"SameCaption", "https://www.example.com", "www.example.com", 0, 1, true, link},
		{"Couple", "https://www.example.com", "", 0, 2, true, couple},
		{"Post", "https://www.site.com", "", 100, 1, true, post},
		{"Empty", "", "", 0, 1, false, empty},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := New()

			var isAdded bool
			for i := 0; i < tt.multiplier; i++ {
				isAdded = result.Add(tt.link, tt.caption, tt.post)
			}

			if isAdded != tt.isAdded {
				t.Fatalf("Получен результат: %v, ожидается: %v", isAdded, tt.isAdded)
			}
			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %#v, ожидается: %#v", result, tt.result)
			}
		})
	}
}

func TestIsExist(t *testing.T) {
	link := New()
	link.Add("https://www.example.com", "", 0)

	tests := []struct {
		test   string
		value  Links
		key    string
		result bool
	}{
		{"Valid", link, "www.example.com", true},
		{"AlsoValid", link, "https://www.example.com/", true},
		{"NotFound", link, "www.site.ru", false},
		{"NoKey", link, "", false},
		{"Empty", Links{}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := tt.value.IsExist(tt.key)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	link := New()
	link.Add("https://www.example.com", "", 0)

	tests := []struct {
		test   string
		value  Links
		key    string
		result Links
	}{
		{"Valid", link, "www.example.com", Links{}},
		{"AlsoValid", link, "https://www.example.com/", Links{}},
		{"NotFound", link, "www.site.ru", link},
		{"NoKey", link, "", link},
		{"Empty", Links{}, "", Links{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := New()
			for k, v := range tt.value {
				result[k] = v
			}

			result.Remove(tt.key)

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %#v, ожидается: %#v", result, tt.result)
			}
		})
	}
}

func TestString(t *testing.T) {
	link := New()
	link.Add("https://www.example.com", "", 0)

	caption := New()
	caption.Add("https://www.example.ru", "Example", 0)

	captions := New()
	captions.Add("https://www.site.io/", "www.site.io", 0)
	captions.Add("https://www.site.io/", "Caption", 0)
	captions.Add("https://www.site.io/", "Ext caption", 0)

	long := New()
	long.Add("https://www.example.com/verry/verryy/long", "", 0)

	empty := New()

	tests := []struct {
		test   string
		value  Links
		result string
	}{
		{"Link", link, "https://www.example.com"},
		{"Caption", caption, "https://www.example.ru [Example]"},
		{"Captions", captions, "https://www.site.io/ [Caption] (3)"},
		{"Long", long, "https://www.example.com/verry/..."},
		{"Empty", empty, ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := tt.value.String()
			if !strings.Contains(result, tt.result) {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	link := New()
	link.Add("https://www.example.com", "", 0)

	captions := New()
	captions.Add("https://www.site.io/", "www.site.io", 0)
	captions.Add("https://www.site.io/", "Caption", 0)

	empty := New()

	tests := []struct {
		test    string
		value   Links
		caption string
		result  string
	}{
		{"Link", link, "Link:", "https://www.example.com"},
		{"Captions", captions, "Links:", "https://www.site.io/ [Caption] (2)"},
		{"Empty", empty, "", ""},
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

func TestClear(t *testing.T) {
	link := New()
	link.Add("https://www.example.com", "", 0)

	tests := []struct {
		test   string
		value  Links
		result Links
	}{
		{"Valid", link, Links{}},
		{"Empty", Links{}, Links{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := New()
			for k, v := range tt.value {
				result[k] = v
			}

			result.Clear()

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %#v, ожидается: %#v", result, tt.result)
			}
		})
	}
}
