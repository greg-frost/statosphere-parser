package cache

import (
	"strings"
	"sync"
	"time"
)

// Кэш
type cache struct {
	isEnabled bool
	values    map[string]cacheData
	success   int
	failed    int
	cancel    chan bool
	m         sync.RWMutex
}

// Данные кэша
type cacheData struct {
	value   interface{}
	expires time.Time
}

// Объект кэша
var c = cache{values: make(map[string]cacheData), cancel: make(chan bool)}

// Включение кэша
func Enable() {
	c.isEnabled = true
}

// Выключение кэша
func Disable() {
	c.isEnabled = false
}

// Запись значения кэша
func SetValue(key string, value interface{}, expires time.Duration) bool {
	c.m.Lock()
	defer c.m.Unlock()

	if !c.isEnabled {
		return false
	}

	key = strings.ToLower(key)

	if key == "" {
		return false
	}

	c.values[key] = cacheData{
		value:   value,
		expires: time.Now().Add(expires),
	}

	return true
}

// Чтение значения кэша
func Value(key string) (interface{}, bool) {
	c.m.Lock()
	defer c.m.Unlock()

	var value interface{}

	if !c.isEnabled {
		return value, false
	}

	key = strings.ToLower(key)
	data, ok := c.values[key]

	if !ok {
		c.failed++
		return value, false
	}

	if time.Now().After(data.expires) {
		c.failed++
		delete(c.values, key)
		return value, false
	}

	c.success++
	return data.value, true
}

// Удаление значения кэша
func Remove(key string) {
	c.m.Lock()
	defer c.m.Unlock()

	if !c.isEnabled {
		return
	}

	key = strings.ToLower(key)
	delete(c.values, key)
}

// Очистка кэша
func Clear() {
	c.m.Lock()
	defer c.m.Unlock()

	c.values = make(map[string]cacheData)
	c.success = 0
	c.failed = 0
}

// Проверка кэша
func Check() {
	c.m.Lock()
	defer c.m.Unlock()

	if !c.isEnabled {
		return
	}

	for key, data := range c.values {
		if time.Now().After(data.expires) {
			delete(c.values, key)
		}
	}
}

// Периодическая проверка кэша
func CheckEvery(interval time.Duration) func() {
	go func() {
		for {
			select {
			case <-time.After(interval):
				Check()
			case <-c.cancel:
				return
			}
		}
	}()

	return func() {
		c.cancel <- true
	}
}

// Статистика кэша
func Stats() (int, int, int) {
	c.m.RLock()
	defer c.m.RUnlock()

	return len(c.values), c.success, c.failed
}
