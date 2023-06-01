package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"statosphere/parser/cache"
	"statosphere/parser/file"
	"statosphere/parser/format"
	"statosphere/parser/get"
	"statosphere/parser/page"
	"statosphere/parser/parse"
	"statosphere/parser/proxy"
)

func main() {
	fmt.Println(" \n[ ПАРСЕР ]\n ")

	// Смена рабочей папки
	file.ChangeDir("parser")

	// Параметры

	var (
		mode, peer, peers, server                     string
		offset, limit, messages, port                 uint
		isConsole, isServer, isProxy, isExact, isTest bool
	)

	flag.StringVar(&mode, "mode", "server", "режим работы (console, server)")
	flag.BoolVar(&isConsole, "console", false, "режим работы в консоле")
	flag.BoolVar(&isServer, "server", false, "режим работы в виде сервера")
	flag.StringVar(&server, "address", "localhost", "адрес сервера")
	flag.UintVar(&port, "port", 8080, "порт сервера")
	flag.BoolVar(&isProxy, "proxy", true, "включение прокси")
	flag.StringVar(&peer, "channel", "", "адрес канала")
	flag.StringVar(&peers, "channels", "", "адреса каналов (через запятую)")
	flag.UintVar(&offset, "offset", 0, "смещение выборки каналов")
	flag.UintVar(&limit, "limit", 100, "ограничение выборки каналов")
	flag.BoolVar(&isExact, "exact", true, "точное число подписчиков")
	flag.UintVar(&messages, "messages", 0, "количество сообщений канала")
	flag.BoolVar(&isTest, "test", false, "тестовый режим (с подборкой каналов)")
	flag.Parse()

	if isConsole {
		mode = "console"
	}
	if isServer {
		mode = "server"
	}

	// Прокси
	if isProxy {
		proxy.Enable(300, 50, 65*time.Second)
		proxy.PrepareFromFile("data/proxies")
	}

	// Режим работы
	switch mode {

	// Сервер
	case "server":
		// Обработчики
		http.HandleFunc("/", page.Index)
		http.HandleFunc("/info", page.Info)
		http.HandleFunc("/messages", page.Messages)
		http.HandleFunc("/styles.css", page.Styles)

		// Подготовка шаблонов
		page.SetTemplate("index", "templates/page.html", "templates/index.html")
		page.SetTemplate("info", "templates/page.html", "templates/info.html", "templates/form.html")
		page.SetTemplate("messages", "templates/page.html", "templates/messages.html", "templates/form.html")

		// Выбор транспорта
		get.SetTransport("http", 0)

		// Включение кэша
		cache.Enable()
		cache.CheckEvery(time.Minute)

		// "Разогрев" http
		go get.Page("https://telegram.org")

		serverAddr := server + ":" + fmt.Sprint(port)
		serverLink := format.Link(serverAddr)

		// Подсказка
		fmt.Println("Сервер ожидает подключений...")
		fmt.Println("(на " + serverLink + ")")

		// Запуск сервера
		http.ListenAndServe(serverAddr, nil)

	// Консоль
	case "console":
		fallthrough
	default:
		// Инициализация
		start := time.Now()
		channels := parse.NewChannels()

		// Заполнение
		switch {
		case peer != "":
			channels.Add(peer)
		case len(peers) > 0:
			channels.PrepareFromString(peers)
		case isTest:
			channels.PrepareFromFile("data/channels")
		}
		channels.Limit(offset, offset+limit)

		// Выбор транспорта
		get.SetTransport("curl", channels.Count())

		// Парсинг
		_, errs := channels.Parse(context.Background(), isExact, messages)

		// Печать
		channels.Print(true)

		// Отчет
		channels.PrintReport(errs, false)

		// Время выполнения
		fmt.Println(time.Now().Sub(start))

	// Прокси
	case "proxy":
		// Инициализация
		channels := parse.NewChannels()

		// Заполнение
		channels.PrepareFromFile("data/channels")

		// Выбор транспорта
		get.SetTransport("http", 0)

		// Тест прокси
		channels.TestProxy(50, 1000, 5*time.Second)
	}
}
