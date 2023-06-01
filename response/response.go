package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Структура ответа
type Response struct {
	Ok        bool          `json:"ok"`
	Code      int           `json:"code"`
	Status    string        `json:"status"`
	Data      interface{}   `json:"data,omitempty"`
	Error     error         `json:"-"`
	ErrorMsg  string        `json:"error,omitempty"`
	Errors    []error       `json:"-"`
	ErrorsMsg []string      `json:"errors,omitempty"`
	Time      time.Duration `json:"-"`
	TimeMsg   string        `json:"time,omitempty"`
}

// Конструктор ответа
func New(data interface{}, count int, errs []error, time time.Duration) Response {
	ok := true
	code := http.StatusOK
	status := http.StatusText(code)

	if count == 0 {
		ok = false
		code = http.StatusNotFound
		status = http.StatusText(code)
	}

	return Response{
		Ok:     ok,
		Code:   code,
		Status: status,
		Data:   data,
		Errors: errs,
		Time:   time,
	}
}

// Приведение некоторых значений ответа к строкам
func (r *Response) ToStrings() {
	if r.Error != nil {
		r.ErrorMsg = fmt.Sprint(r.Error)
	}

	r.ErrorsMsg = make([]string, 0)
	for _, err := range r.Errors {
		r.ErrorsMsg = append(r.ErrorsMsg, fmt.Sprint(err))
	}

	r.TimeMsg = fmt.Sprint(r.Time)
}

// Возвращение значений некоторым строкам ответа
func (r *Response) ToValues() {
	if r.ErrorMsg != "" {
		r.Error = errors.New(r.ErrorMsg)
	}

	r.Errors = make([]error, 0)
	for _, err := range r.ErrorsMsg {
		r.Errors = append(r.Errors, errors.New(err))
	}

	r.Time, _ = time.ParseDuration(r.TimeMsg)
}

// Кодирование JSON
func (r *Response) EncodeJSON() (string, error) {
	r.ToStrings()

	result, err := json.Marshal(r)
	return string(result), err
}

// Декодирование JSON
func (r *Response) DecodeJSON(text string) error {
	err := json.Unmarshal([]byte(text), r)

	r.ToValues()
	return err
}

// Печать JSON
func PrintJSON(res http.ResponseWriter, json string) {
	res.Header().Set("Content-Type", "application/json")
	fmt.Fprint(res, json)
}

// Печать кода и служебного текста
func PrintStatus(res http.ResponseWriter, code int) {
	res.WriteHeader(code)
	fmt.Fprint(res, code, " ", http.StatusText(code))
}
