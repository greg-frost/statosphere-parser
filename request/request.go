package request

import (
	"net/http"
	"sync"
	"time"

	"statosphere/parser/check"
)

// Запрос
type request struct {
	users map[string]*user
	m     sync.Mutex
}

// Пользователь запроса
type user struct {
	requests     uint
	firstRequest time.Time
}

// Глобальный объект запроса
var r = request{users: make(map[string]*user)}

// Ограничение запроса
func Barrier(key string, requests uint, duration time.Duration) bool {
	r.m.Lock()
	defer r.m.Unlock()

	u, ok := r.users[key]
	if !ok {
		r.users[key] = &user{}
		u = r.users[key]
		u.firstRequest = time.Now()
	}

	u.requests++

	if time.Now().After(u.firstRequest.Add(duration)) {
		delete(r.users, key)
		return true
	}

	if u.requests > requests {
		return false
	}

	return true
}

// Получение числового поля запроса
func ParamInt(req *http.Request, fieldName string, defaultValue int) int {
	res, err := check.Int(req.FormValue(fieldName))
	if err != nil {
		res = defaultValue
	}

	return res
}

// Получение положительного числового поля запроса
func ParamPositiveInt(req *http.Request, fieldName string, defaultValue uint) uint {
	res, err := check.PositiveInt(req.FormValue(fieldName))
	if err != nil {
		res = defaultValue
	}

	return res
}

// Получение логического поля запроса
func ParamBool(req *http.Request, fieldName string, defaultValue bool) bool {
	res, err := check.Bool(req.FormValue(fieldName))
	if err != nil {
		res = defaultValue
	}

	return res
}

// Получение строкового поля запроса
func ParamString(req *http.Request, fieldName string, defaultValue string) string {
	res := req.FormValue(fieldName)
	if res == "" {
		res = defaultValue
	}

	return res
}

// Получение списка поля запроса
func ParamList(req *http.Request, fieldName string, defaultValue []string) []string {
	req.ParseForm()
	if list, ok := req.Form[fieldName+"[]"]; ok {
		return list
	}

	str := req.FormValue(fieldName)
	if str == "" {
		return defaultValue
	}

	list, _ := check.List(str)

	return list
}
