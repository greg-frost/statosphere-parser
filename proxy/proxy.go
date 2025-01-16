package proxy

import (
	"net/url"
	"strings"
	"sync"
	"time"

	"statosphere/parser/file"
)

// Прокси
type proxy struct {
	isEnabled      bool
	proxies        []string
	current        int
	last           int
	requests       int
	mainThreshold  int
	proxyThreshold int
	cooldown       time.Duration
	lastCooldown   time.Time
	m              sync.Mutex
}

// Объект прокси
var p = proxy{proxies: []string{""}}

// Включение прокси
func Enable(mainThreshold, proxyThreshold uint, cooldown time.Duration) {
	p.isEnabled = true

	p.mainThreshold = int(mainThreshold)
	p.proxyThreshold = int(proxyThreshold)

	p.cooldown = cooldown
	p.lastCooldown = time.Now()
}

// Выключение прокси
func Disable() {
	p.isEnabled = false
}

// Добавление прокси
func Prepare(list []string) {
	proxies := make([]string, 0, len(list))

	for _, proxy := range list {
		proxy = strings.TrimSpace(proxy)

		if proxy == "" {
			continue
		}

		if !strings.HasPrefix(proxy, "http://") && !strings.HasPrefix(proxy, "https://") {
			proxy = "http://" + proxy
		}

		_, err := url.ParseRequestURI(proxy)
		if err != nil {
			continue
		}

		proxies = append(proxies, proxy)
	}

	if len(proxies) == 0 {
		return
	}

	p.proxies = append(p.proxies, proxies...)
}

// Добавление прокси из списка
func PrepareFromList(list []string) {
	Prepare(list)
}

// Добавление прокси из файла
func PrepareFromFile(filename string) {
	list, _ := file.ReadLines(filename)
	Prepare(list)
}

// Сброс прокси
func Reset() {
	p.current = 0
	p.last = 0
	p.requests = 0

	p.lastCooldown = time.Time{}
}

// Очистка прокси
func Clear() {
	p.proxies = []string{""}
	Reset()
}

// Получение адреса текущего прокси
func Current() string {
	p.m.Lock()
	defer p.m.Unlock()

	var proxy string

	if !p.isEnabled {
		return proxy
	}

	proxy = p.proxies[p.current]

	return proxy
}

// Получение адреса следующего прокси
func Next() string {
	p.m.Lock()
	defer p.m.Unlock()

	var proxy string

	if !p.isEnabled {
		return proxy
	}

	proxy = p.proxies[p.current]

	p.last = p.current
	p.current = (p.current + 1) % len(p.proxies)
	p.requests = 0

	return proxy
}

// Обход прокси по кругу
func Round() string {
	p.m.Lock()
	defer p.m.Unlock()

	var proxy string

	if !p.isEnabled {
		return proxy
	}

	p.requests++

	if p.current == 0 && p.requests > p.mainThreshold || p.current > 0 && p.requests > p.proxyThreshold {
		p.current = (p.current + 1) % len(p.proxies)
		p.requests = 1
	}

	proxy = p.proxies[p.current]

	return proxy
}

// Обход прокси по кругу (с кулдауном)
func Cooldown() string {
	p.m.Lock()
	defer p.m.Unlock()

	var proxy string

	if !p.isEnabled {
		return proxy
	}

	p.requests++

	if p.current == 0 && p.requests > p.mainThreshold || p.current > 0 && p.requests > p.proxyThreshold {
		p.requests = 1

		if p.last == 0 {
			p.current = (p.current + 1) % len(p.proxies)
			if p.current == 0 {
				p.current = (p.current + 1) % len(p.proxies)
			}
		} else {
			p.current = p.last
			p.last = 0
		}

		if time.Now().After(p.lastCooldown.Add(p.cooldown)) {
			p.lastCooldown = time.Now()

			p.last = p.current
			p.current = 0
		}
	}

	proxy = p.proxies[p.current]

	return proxy
}
