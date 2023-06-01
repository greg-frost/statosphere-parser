package links

import (
	"fmt"
	"strings"

	"statosphere/parser/format"
)

// Ссылка
type Link struct {
	Link     string   `json:"link"`
	Captions []string `json:"captions,omitempty"`
	Posts    []uint   `json:"posts,omitempty"`
	Pos      int      `json:"pos"`
	Count    int      `json:"count"`
}

// Ссылки
type Links map[string]Link

// Конструктор набора ссылок
func New() Links {
	return make(Links)
}

// Формирование ключа
func Key(link string) string {
	if link == "" {
		return link
	}

	key, _, _ := strings.Cut(link, "?")
	key = strings.ToLower(key)
	key = format.TrimLink(key)
	key = format.Truncate(key, 32)

	return key
}

// Добавление ссылки
func (l Links) Add(link, caption string, post uint) bool {
	if link == "" {
		return false
	}

	key := Key(link)

	if l[key].Link != "" {
		link = l[key].Link
	}

	captions := l[key].Captions
	posts := l[key].Posts
	pos := l[key].Pos

	caption = format.SafeString(caption)

	if len(caption) >= 3 {
		clearLink := format.TrimLink(link)
		clearCaption := format.TrimLink(caption)

		if link == caption || clearLink == clearCaption {
			caption = ""
		}
	} else {
		caption = ""
	}

	if caption != "" {
		captions = append(captions, caption)
	}

	if post != 0 {
		posts = append(posts, post)
	}

	if pos == 0 {
		pos = len(l) + 1
	}

	l[key] = Link{
		Link:     link,
		Captions: captions,
		Posts:    posts,
		Pos:      pos,
		Count:    l[key].Count + 1,
	}

	return true
}

// Проверка существования ссылки
func (l Links) IsExist(link string) bool {
	if link == "" {
		return false
	}

	key := Key(link)
	_, ok := l[key]

	return ok
}

// Приведение ссылок к строке
func (l Links) String() string {
	var res string
	for _, data := range l {
		res += "\n" + "   " + format.Truncate(data.Link, 30)
		if len(data.Captions) > 0 {
			res += fmt.Sprintf(" [%s]", format.Truncate(data.Captions[0], 20))
		}
		if data.Count > 1 {
			res += fmt.Sprintf(" (%d)", data.Count)
		}
	}
	return res
}

// Печать ссылок
func (l Links) Print(title string) {
	if len(l) == 0 {
		return
	}

	fmt.Println(title, l)
}

// Удаление ссылки
func (l Links) Remove(link string) {
	if link == "" {
		return
	}

	key := Key(link)
	delete(l, key)
}

// Удаление всех ссылок
func (l *Links) Clear() {
	*l = make(Links)
}
