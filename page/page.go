package page

import (
	"html/template"
	"net/http"
	"sync"
	"time"

	"statosphere/parser/parse"
	"statosphere/parser/request"
	"statosphere/parser/response"
)

// Шаблоны
type templates struct {
	templates map[string]*template.Template
	m         sync.RWMutex
}

// Объект шаблонов
var t = templates{templates: make(map[string]*template.Template)}

// Данные шаблона страницы
type templatePage struct {
	Title, Caption string
}

// Данные шаблона меню
type templateMenu struct {
	Page    templatePage
	Buttons []templateButton
}

// Данные шаблона формы
type templateForm struct {
	Page   templatePage
	Fields map[string]templateField
	Submit string
}

// Данные шаблона кнопки
type templateButton struct {
	Link, Caption string
}

// Данные шаблона поля
type templateField struct {
	Caption string
	Value   interface{}
}

// Сохранение шаблона
func SetTemplate(key string, files ...string) {
	t.m.Lock()
	defer t.m.Unlock()

	tpl, err := template.ParseFiles(files...)
	if err != nil {
		tpl, _ = template.New(key).Parse("Веб-форма недоступна, используйте REST")
	}

	t.templates[key] = tpl
}

// Получение шаблона
func Template(key string) *template.Template {
	t.m.RLock()
	defer t.m.RUnlock()

	return t.templates[key]
}

// Главная страница
func Index(res http.ResponseWriter, req *http.Request) {
	p := templateMenu{
		Page: templatePage{
			Title:   "Парсер",
			Caption: "Выберите режим работы:",
		},
		Buttons: []templateButton{
			{Link: "info", Caption: "Информация"},
			{Link: "messages", Caption: "Сообщения"},
		},
	}

	Template("index").Execute(res, p)
}

// Страница с информацией каналов
func Info(res http.ResponseWriter, req *http.Request) {

	// Ограничение числа запросов
	if !request.Barrier("info", 10, time.Duration(time.Second)) {
		response.PrintStatus(res, http.StatusTooManyRequests)
		return
	}

	// Форма при пустом запросе
	if req.Method == "GET" && req.URL.RawQuery == "" {
		p := templateForm{
			Page: templatePage{
				Title: "Информация",
				Caption: `Введите ссылки на ТГ-каналы через запятую
					или выберите режим "Тестовый набор"`,
			},
			Fields: map[string]templateField{
				"Channels": {Caption: "Каналы"},
				"Test":     {Caption: "Тестовый набор", Value: true},
				"Offset":   {Caption: "Смещение", Value: 0},
				"Limit":    {Caption: "Лимит", Value: 5},
			},
			Submit: "Начать парсинг",
		}

		Template("info").Execute(res, p)
		return
	}

	// Инициализация
	start := time.Now()
	channels := parse.NewChannels()

	// Параметры
	peer := request.ParamString(req, "channel", "")
	peers := request.ParamList(req, "channels", []string{})
	limit := request.ParamPositiveInt(req, "limit", 100)
	offset := request.ParamPositiveInt(req, "offset", 0)
	isTest := request.ParamBool(req, "test", false)

	// Заполнение
	switch {
	case peer != "":
		channels.Add(peer)
	case len(peers) > 0:
		channels.PrepareFromList(peers)
	case isTest:
		channels.PrepareFromFile("data/channels")
	}
	channels.Limit(offset, offset+limit)

	// Парсинг
	_, errs := channels.Parse(req.Context(), true, 0)
	channels.RemoveUnparsed()

	// Подготовка JSON
	json := response.New(
		channels.Channels.Channels, channels.Count(),
		errs, time.Since(start),
	)
	result, _ := json.EncodeJSON()

	// Печать JSON
	response.PrintJSON(res, result)
}

// Страница с сообщениями каналов
func Messages(res http.ResponseWriter, req *http.Request) {

	// Ограничение числа запросов
	if !request.Barrier("msgs", 5, time.Duration(time.Second)) {
		response.PrintStatus(res, http.StatusTooManyRequests)
		return
	}

	// Форма при пустом запросе
	if req.Method == "GET" && req.URL.RawQuery == "" {
		p := templateForm{
			Page: templatePage{
				Title: "Сообщения",
				Caption: `Введите ссылки на ТГ-каналы через запятую или 
					выберите режим "Тестовый набор"`,
			},
			Fields: map[string]templateField{
				"Channels": {Caption: "Каналы"},
				"Test":     {Caption: "Тестовый набор", Value: true},
				"Offset":   {Caption: "Смещение", Value: 0},
				"Limit":    {Caption: "Лимит", Value: 5},
				"Messages": {Caption: "Сообщений", Value: 10},
				"Exact":    {Caption: "Точное число подписчиков"},
			},
			Submit: "Начать парсинг",
		}

		Template("messages").Execute(res, p)
		return
	}

	// Инициализация
	start := time.Now()
	channels := parse.NewChannels()

	// Параметры
	peer := request.ParamString(req, "channel", "")
	peers := request.ParamList(req, "channels", []string{})
	limit := request.ParamPositiveInt(req, "limit", 100)
	offset := request.ParamPositiveInt(req, "offset", 0)
	messages := request.ParamPositiveInt(req, "messages", 20)
	isExact := request.ParamBool(req, "exact", false)
	isTest := request.ParamBool(req, "test", false)

	// Заполнение
	switch {
	case peer != "":
		channels.Add(peer)
	case len(peers) > 0:
		channels.PrepareFromList(peers)
	case isTest:
		channels.PrepareFromFile("data/channels")
	}
	channels.Limit(offset, offset+limit)

	// Парсинг
	_, errs := channels.Parse(req.Context(), isExact, messages)
	channels.RemoveUnmessaged()

	// Подготовка JSON
	json := response.New(
		channels.Channels.Channels, channels.Count(),
		errs, time.Since(start),
	)
	result, _ := json.EncodeJSON()

	// Печать JSON
	response.PrintJSON(res, result)
}

// Страница со стилями
func Styles(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "templates/styles.css")
}
