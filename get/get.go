package get

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"

	"statosphere/parser/check"
	"statosphere/parser/file"
	"statosphere/parser/proxy"
)

var (
	CurlFasterTreshold = 50    // Предел запросов, до которого curl может оказаться быстрее http
	IsCacheDisable     = false // Отключение кэша
)

// Получение страницы (оптимальным способом)
func Page(address string) (int, string, error) {
	proxy := proxy.Cooldown()

	switch Transport() {
	case "http":
		fallthrough
	default:
		return PageHTTP(address, proxy)
	case "curl":
		return PageCURL(address, proxy)
	case "file":
		return PageFile(address, proxy)
	}
}

// Получение страницы через пакет http
func PageHTTP(address, proxy string) (int, string, error) {
	var code = 500

	if address == "" {
		return code, "", errors.New("url-адрес пуст")
	}

	req, _ := http.NewRequest("GET", address, nil)

	client := &http.Client{}

	// Таймаут
	timeout := Timeout()

	if timeout > 0 {
		client.Timeout = timeout
	}

	// Прокси

	if proxy != "" {
		proxyUrl, err := url.ParseRequestURI(proxy)
		if err != nil {
			return code, "", fmt.Errorf("url-адрес прокси %q не валиден", proxy)
		}

		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	}

	// Отключение кэша
	if IsCacheDisable {
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Pragma", "no-cache")
	}

	// Запрос

	resp, err := client.Do(req)
	if resp != nil {
		code = resp.StatusCode
	}
	if err != nil {
		return code, "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return code, string(body), nil
}

// Получение страницы через внешний Curl
func PageCURL(address, proxy string) (int, string, error) {
	var code = 500

	if address == "" {
		return code, "", errors.New("url-адрес пуст")
	}

	var params []string

	// Игнорирование ssl-сертификатов
	params = append(params, "-k")

	// Таймаут
	timeout := TimeoutString()

	if timeout != "" {
		params = append(params, "-m")
		params = append(params, timeout)
	}

	// Прокси
	if proxy != "" {
		params = append(params, "-x")
		params = append(params, proxy)
	}

	// Отключение кэша
	if IsCacheDisable {
		params = append(params, "-H")
		params = append(params, "Cache-Control: no-cache")

		params = append(params, "-H")
		params = append(params, "Pragma: no-cache")
	}

	params = append(params, address)

	// Запрос

	curl := exec.Command("curl", params...)

	body, err := curl.Output()
	if err != nil {
		switch e := err.(type) {
		case *exec.ExitError:
			switch e.ExitCode() {
			case 28:
				err = fmt.Errorf("превышен таймаут (%v сек.)", timeout)
			case 5, 56:
				err = errors.New("прокси невалиден (недоступен)")
			}
		}
		return code, "", err
	}

	return 200, string(body), nil
}

// Получение страницы из файла
func PageFile(address, proxy string) (int, string, error) {
	var code = 500

	if address == "" {
		return code, "", errors.New("url-адрес пуст")
	}

	path := "data/pages/"

	switch strings.Contains(address, "/s/") {
	case false:
		path += "info/"
	case true:
		path += "msgs/"
	}

	username, joinchat, _ := check.Username(address, true)

	switch {
	case username != "":
		path += username
	case joinchat != "":
		path += joinchat
	default:
		return code, "", errors.New("url-адрес невалиден")
	}

	code = 200

	body, err := file.Read(path)
	if err != nil {
		code, body, err = PageCURL(address, proxy)
		if body != "" {
			file.Write(path, body)
		}
	}

	return code, body, err
}

// Опции
type options struct {
	transport string
	timeout   time.Duration
	m         sync.RWMutex
}

// Объект опций
var o = options{transport: "http", timeout: 15 * time.Second}

// Запись нового значения транспорта
func SetTransport(transport string, requests int) {
	o.m.Lock()
	defer o.m.Unlock()

	transport = strings.ToLower(transport)

	o.transport = "http"
	switch {
	case transport == "curl" && requests <= CurlFasterTreshold:
		o.transport = "curl"
	case transport == "file":
		o.transport = "file"
	}
}

// Получение текущего значения транспорта
func Transport() string {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.transport
}

// Запись нового значения таймаута
func SetTimeout(timeout time.Duration) {
	o.m.Lock()
	defer o.m.Unlock()

	o.timeout = timeout
}

// Получение текущего значения таймаута
func Timeout() time.Duration {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.timeout
}

// Получение текущего значения таймаута (в виде строки)
func TimeoutString() string {
	o.m.RLock()
	defer o.m.RUnlock()

	timeout := fmt.Sprintf("%0.2f", float64(o.timeout)/float64(time.Second))
	timeout = strings.ReplaceAll(timeout, ".", ",")
	timeout = strings.ReplaceAll(timeout, ",00", "")

	return timeout
}
