package page

import (
	"html/template"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"statosphere/parser/file"
	"statosphere/parser/get"
	"statosphere/parser/mock"
)

func init() {
	os.Chdir("..")
	get.SetTransport("file", 0)
}

var ts = t.templates

func TestSetTemplate(t *testing.T) {
	one := template.Must(template.ParseFiles("templates/page.html"))
	many := template.Must(template.ParseFiles("templates/info.html", "templates/form.html"))

	tests := []struct {
		test   string
		key    string
		value  []string
		result *template.Template
	}{
		{"One", "one", []string{"templates/page.html"}, one},
		{"Many", "many", []string{"templates/info.html", "templates/form.html"}, many},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			SetTemplate(tt.key, tt.value...)
			result := ts[tt.key]

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestTemplate(t *testing.T) {
	ts["one"] = template.Must(template.ParseFiles("templates/page.html"))
	ts["many"] = template.Must(template.ParseFiles("templates/info.html", "templates/form.html"))

	tests := []struct {
		test   string
		key    string
		result *template.Template
	}{
		{"One", "one", ts["one"]},
		{"Many", "many", ts["many"]},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Template(tt.key)

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestIndex(t *testing.T) {
	SetTemplate("index", "templates/page.html", "templates/index.html")

	tests := []struct {
		test   string
		result []string
	}{
		{"Valid", []string{"!DOCTYPE", "<html>", "<head>", "utf-8", "<title>",
			"styles.css", "<body>", `href="info"`, `href="messages"`}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			response := mock.NewResponseWriter()
			request := mock.NewRequest()

			Index(response, &request.Request)

			result := response.Body()

			for _, match := range tt.result {
				if !strings.Contains(result, match) {
					t.Fatalf("Получено значение: %v, ожидаются совпадения: %v, не найдено: %v",
						result, tt.result, match)
				}
			}
		})
	}
}

func TestInfo(t *testing.T) {
	SetTemplate("info", "templates/page.html", "templates/info.html", "templates/form.html")

	tests := []struct {
		test   string
		value  map[string]interface{}
		times  int
		result []string
	}{
		{"Form", nil, 1, []string{"!DOCTYPE", "<html>", "<head>", "utf-8", "<title>", "styles.css", "<body>",
			`form class="form"`, `input type="text" name="channels"`, `input type="checkbox" name="test"`,
			`input type="text" name="offset"`, `input type="text" name="limit"`, `input type="submit"`}},
		{"Channel", map[string]interface{}{"channel": "thecodemedia"}, 1, []string{
			`"ok":true`, `"code":200`, "data", `"username":"thecodemedia"`, `"peer":"@thecodemedia"`,
			`"link":"https://t.me/thecodemedia"`, "title", `"kind":"channel"`, "participants", "exact"}},
		{"Channels", map[string]interface{}{"channels": "[thecodemedia,codecamp]"}, 1, []string{
			`"code":200`, `"status":"OK"`, "data", `"username":"thecodemedia"`, `"peer":"@thecodemedia"`,
			`"username":"codecamp"`, `"peer":"@codecamp"`, `"kind":"channel"`, "participants"}},
		{"Test", map[string]interface{}{"test": true, "limit": 1}, 1,
			[]string{`"ok":true`, `"code":200`, "data", "kind", "participants"}},
		{"Empty", map[string]interface{}{"limit": 0}, 1,
			[]string{`"ok":false`, `"code":404`, `"status":"Not Found"`, "errors"}},
		{"Flood", map[string]interface{}{"limit": 0}, 15,
			[]string{"429", "Too Many Requests"}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			response := mock.NewResponseWriter()
			request := mock.NewRequest()

			for k, v := range tt.value {
				request.AddParamGET(k, v)
			}

			for i := 0; i < tt.times; i++ {
				response.ClearBody()
				Info(response, &request.Request)
			}

			result := response.Body()

			for _, match := range tt.result {
				if !strings.Contains(result, match) {
					t.Fatalf("Получено значение: %v, ожидаются совпадения: %v, не найдено: %v",
						result, tt.result, match)
				}
			}
		})
	}
}

func TestMessages(t *testing.T) {
	SetTemplate("messages", "templates/page.html", "templates/messages.html", "templates/form.html")

	tests := []struct {
		test   string
		value  map[string]interface{}
		times  int
		result []string
	}{
		{"Form", nil, 1, []string{"!DOCTYPE", "<html>", "<head>", "utf-8", "<title>", "styles.css", "<body>",
			`form class="form"`, `input type="text" name="channels"`, `input type="checkbox" name="test"`,
			`input type="text" name="offset"`, `input type="text" name="limit"`, `input type="text" name="messages"`,
			`input type="checkbox" name="exact"`, `input type="submit"`}},
		{"Channel", map[string]interface{}{"channel": "glav_hack", "exact": true}, 1,
			[]string{`"ok":false`, `"code":404`, `"status":"Not Found"`, "errors"}},
		{"Channels", map[string]interface{}{"channels": "[thecodemedia,glav_hack]", "messages": 5}, 1,
			[]string{`"ok":true`, `"code":200`, "data", `"link":"https://t.me/thecodemedia"`, "title",
				`"kind":"channel"`, "participants", "approx", "messages", "id", "message", "views", "short"}},
		{"Test", map[string]interface{}{"test": true, "messages": 1, "limit": 1}, 1,
			[]string{`"ok":true`, `"code":200`, "data", "kind", "participants"}},
		{"Empty", map[string]interface{}{"limit": 0}, 1,
			[]string{`"ok":false`, `"code":404`, `"status":"Not Found"`, "errors"}},
		{"Flood", map[string]interface{}{"limit": 0}, 10,
			[]string{"429", "Too Many Requests"}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			response := mock.NewResponseWriter()
			request := mock.NewRequest()

			for k, v := range tt.value {
				request.AddParamGET(k, v)
			}

			log.Print(tt.test)
			for i := 0; i < tt.times; i++ {
				response.ClearBody()
				Messages(response, &request.Request)
			}

			result := response.Body()

			for _, match := range tt.result {
				if !strings.Contains(result, match) {
					t.Fatalf("Получено значение: %v, ожидаются совпадения: %v, не найдено: %v",
						result, tt.result, match)
				}
			}

			time.Sleep(50 * time.Millisecond)
		})
	}
}

func TestStyles(t *testing.T) {
	styles, _ := file.Read("templates/styles.css")

	tests := []struct {
		test   string
		result string
	}{
		{"Valid", styles},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			response := mock.NewResponseWriter()
			request := mock.NewRequest()

			Styles(response, &request.Request)
			result := response.Body()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}
