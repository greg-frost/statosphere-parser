package mock

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Ответ-писатель
type ResponseWriter struct {
	headers http.Header
	results map[string]interface{}
}

// Конструктор ответа
func NewResponseWriter() ResponseWriter {
	return ResponseWriter{
		headers: make(http.Header),
		results: make(map[string]interface{}),
	}
}

// Установка заголовка
func (rw ResponseWriter) Header() http.Header {
	return rw.headers
}

// Установка кода
func (rw ResponseWriter) WriteHeader(statusCode int) {
	rw.results["code"] = statusCode
}

// Запись данных
func (rw ResponseWriter) Write(n []byte) (int, error) {
	if rw.results["body"] == nil {
		rw.results["body"] = string(n)
	} else {
		rw.results["body"] = rw.results["body"].(string) + string(n)
	}

	return len(n), nil
}

// Получение заголовка
func (rw ResponseWriter) Head(header string) string {
	return strings.Join(rw.headers[header], ",")
}

// Получение кода
func (rw ResponseWriter) Code() int {
	return rw.results["code"].(int)
}

// Получение тела
func (rw ResponseWriter) Body() string {
	return rw.results["body"].(string)
}

// Очистка тела
func (rw ResponseWriter) ClearBody() {
	rw.results["body"] = ""
}

// Запрос
type Request struct {
	http.Request
	get  url.Values
	post url.Values
}

// Конструктор запроса
func NewRequest() Request {
	return Request{
		Request: http.Request{Method: "GET", URL: &url.URL{}},
		get:     make(url.Values),
		post:    make(url.Values),
	}
}

// Добавление GET-параметра
func (r *Request) AddParamGET(key string, value interface{}) {
	r.Method = "GET"
	r.get.Add(key, fmt.Sprintf("%v", value))
	query, _ := url.Parse("?" + r.get.Encode())
	r.URL = query
}

// Добавление POST-параметра
func (r *Request) AddParamPOST(key string, value interface{}) {
	r.Method = "POST"
	r.post.Add(key, fmt.Sprintf("%v", value))
	r.PostForm = r.post
}

// Перехват и чтение вывода (os.Stdout)
func ReadOutput(printing func()) string {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printing()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	return string(out)
}
