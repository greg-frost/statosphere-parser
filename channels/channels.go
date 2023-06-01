package channels

import (
	"fmt"

	"statosphere/parser/channel"
	"statosphere/parser/check"
	"statosphere/parser/file"
	"statosphere/parser/format"
)

// Набор каналов
type Channels struct {
	Channels []*channel.Channel
}

// Конструктор набора каналов
func New() Channels {
	return Channels{
		Channels: make([]*channel.Channel, 0),
	}
}

// Добавление канала
func (cc *Channels) Add(link string) bool {
	username, joinchat, _ := check.Username(link, false)

	if username == "" && joinchat == "" {
		return false
	}

	c := &channel.Channel{}
	c.Username = username
	c.Joinchat = joinchat
	c.Peer, c.Link = format.Username(username, joinchat, 0)

	cc.Channels = append(cc.Channels, c)

	return true
}

// Подготовка каналов
func (cc *Channels) Prepare(links []string) (n int) {
	for _, link := range links {
		if ok := cc.Add(link); ok {
			n++
		}
	}
	return n
}

// Подготовка каналов из строки
func (cc *Channels) PrepareFromString(str string) int {
	list, _ := check.List(str)
	return cc.Prepare(list)
}

// Подготовка каналов из списка
func (cc *Channels) PrepareFromList(links []string) int {
	return cc.Prepare(links)
}

// Подготовка каналов из файла
func (cc *Channels) PrepareFromFile(filename string) int {
	links, _ := file.ReadLines(filename)
	return cc.Prepare(links)
}

// Поиск канала
func (cc *Channels) Find(link string) (*channel.Channel, bool) {
	var result *channel.Channel
	username, joinchat, _ := check.Username(link, false)

	if username == "" && joinchat == "" {
		return result, false
	}

	for _, c := range cc.Channels {
		if username == c.Username && joinchat == c.Joinchat {
			return c, true
		}
	}

	return result, false
}

// Получение списка каналов
func (cc *Channels) List() []string {
	links := make([]string, 0, cc.Count())

	for _, c := range cc.Channels {
		links = append(links, c.Link)
	}

	return links
}

// Срез нужных каналов
func (cc *Channels) Limit(from, to uint) {
	if to == 0 || from >= to {
		cc.Clear()
		return
	}

	to = uint(check.Max(int(to), cc.Count()))
	cc.Channels = cc.Channels[from:to]
}

// Количество каналов
func (cc *Channels) Count() int {
	return len(cc.Channels)
}

// Печать каналов
func (cc *Channels) Print(isPrintMessages bool) {
	if cc.Count() == 0 {
		return
	}

	fmt.Print("Channels\n--------\n\n")

	for _, c := range cc.Channels {
		c.Print(isPrintMessages)
	}
}

// Печать сообщений каналов
func (cc *Channels) PrintMessages() {
	if cc.Count() == 0 {
		return
	}

	fmt.Print("Messages\n--------\n\n")

	for _, c := range cc.Channels {
		if len(c.Messages) == 0 {
			continue
		}

		fmt.Printf("[ %s ]\n", c.Link)

		c.PrintMessages()

		fmt.Print("\n-----\n\n")
	}
}

// Удаление канала
func (cc *Channels) Remove(link string) bool {
	return cc.RemoveByLink(link)
}

// Удаление канала по индексу
func (cc *Channels) RemoveByIdx(i int) bool {
	if int(i) >= cc.Count() {
		return false
	}

	cc.Channels = append(cc.Channels[:i], cc.Channels[i+1:]...)
	return true
}

// Удаление канала по ссылке
func (cc *Channels) RemoveByLink(link string) bool {
	username, joinchat, _ := check.Username(link, false)

	if username == "" && joinchat == "" {
		return false
	}

	for i, c := range cc.Channels {
		if username == c.Username && joinchat == c.Joinchat {
			return cc.RemoveByIdx(i)
		}
	}

	return false
}

// Удаление дубликатов каналов
func (cc *Channels) RemoveDuplicates() {
	existed := make(map[string]int)
	count := cc.Count()

	for i := 0; i < count; i++ {
		if _, ok := existed[cc.Channels[i].Link]; ok {
			cc.RemoveByIdx(i)
			count--
			i--
		}
		existed[cc.Channels[i].Link]++
	}
}

// Удаление всех каналов
func (cc *Channels) Clear() {
	cc.Channels = make([]*channel.Channel, 0)
}
