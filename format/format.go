package format

import (
	"fmt"
	"html"
	"strings"
	"time"

	"statosphere/parser/regexp"
)

// Формирование юзернейма (джойнчата) канала
func Username(username, joinchat string, post uint) (peer, link string) {
	if username == "" && joinchat == "" {
		return
	}

	switch {
	case username != "":
		peer = "@" + username
		link = "https://t.me/" + username
	case joinchat != "":
		peer = "+" + joinchat
		link = "https://t.me/joinchat/" + joinchat
	}

	if post > 0 {
		peer += "/" + fmt.Sprint(post)
		link += "/" + fmt.Sprint(post)
	}

	return
}

// Формирование веб-ссылки канала
func PageLink(username, joinchat string) (info, messages string) {
	if username == "" && joinchat == "" {
		return
	}

	switch {
	case username != "":
		info = "https://t.me/" + username
		messages = "https://t.me/s/" + username
	case joinchat != "":
		info = "https://t.me/joinchat/" + joinchat
		messages = ""
	}

	return
}

// Сокращение страницы
func StripPage(text string) string {
	if text == "" {
		return text
	}

	text = StripSegments(text, "<head>", "</header>")
	text = StripSegments(text, "<svg", "</svg>")
	text = StripSegments(text, "tgme_widget_message_user", "</div>")

	return text
}

// Сокращение страницы (regexp)
func StripPageRegexp(text string) string {
	if text == "" {
		return text
	}

	re := regexp.Prepare("stripHead", `(?s)<head>.+?</header>`)
	text = re.ReplaceAll(text, "")

	re = regexp.Prepare("stripSvg", `(?s)<svg[^>]*?>.+?</svg>`)
	text = re.ReplaceAll(text, "")

	re = regexp.Prepare("stripAvatar", `(?s)tgme_widget_message_user[^<]+?</div>`)
	text = re.ReplaceAll(text, "")

	return text
}

// Вырезание сегментов
func StripSegments(text, open, close string) string {
	if text == "" || open == "" || close == "" {
		return text
	}

	var (
		start, end int
		search     string
	)

	for {
		search = text

		start = strings.Index(search, open)
		if start == -1 {
			break
		}
		search = search[start:]

		end = strings.Index(search, close)
		if end == -1 {
			break
		}
		text = text[:start] + text[start+end+len(close):]
	}

	return text
}

// Получение безопасного HTML
func SafeHtml(text string) string {
	if text == "" {
		return text
	}

	text = Unescape(text)
	text = Newlines(text)
	text = StripHidden(text)
	text = Clean(text)
	text = Trim(text)

	return text
}

// Получение безопасной строки
func SafeString(text string) string {
	if text == "" {
		return text
	}

	text = Unescape(text)
	text = StripNewlines(text)
	text = StripTags(text)
	text = StripHiddenRegexp(text)
	text = Trim(text)

	return text
}

// Получение безопасного текста
func SafeText(text string) string {
	if text == "" {
		return text
	}

	text = Unescape(text)
	text = Newlines(text)
	text = StripTags(text)
	text = StripHidden(text)
	text = Trim(text)

	return text
}

// Замена спецсимволов
func Unescape(text string) string {
	if text == "" {
		return text
	}

	text = html.UnescapeString(text)
	text = strings.ReplaceAll(text, "&amp;", "&")

	return text
}

// Замена переносов строк
func Newlines(text string) string {
	if text == "" {
		return text
	}

	re := regexp.Prepare("formatNewlines", `(\t|<br.*?>)`)
	text = re.ReplaceAll(text, "\n")

	return text
}

// Удаление переносов строк
func StripNewlines(text string) string {
	if text == "" {
		return text
	}

	re := regexp.Prepare("stripNewlines", `(\r?\n\r?|<br.*?>)`)
	text = re.ReplaceAll(text, "")

	return text
}

// Удаление тегов
func StripTags(text string) string {
	if text == "" {
		return text
	}

	re := regexp.Prepare("stripOpenerTags", `<.*?>`)
	text = re.ReplaceAll(text, "")

	re = regexp.Prepare("stripCloserTags", `</.*?>`)
	text = re.ReplaceAll(text, "")

	return text
}

// Удаление служебных тегов
func StripServiceTags(text string) string {
	if text == "" {
		return text
	}

	text = StripCode(text)
	text = StripSpoilers(text)
	//text = StripEmptyLinks(text)

	return text
}

// Удаление тегов с кодом
func StripCode(text string) string {
	if text == "" {
		return text
	}

	re := regexp.Prepare("stripCode", `(?s)<code>(.+?)</code>`)
	text = re.ReplaceAll(text, "")

	re = regexp.Prepare("stripPre", `(?s)<pre>(.+?)</pre>`)
	text = re.ReplaceAll(text, "")

	return text
}

// Удаление спойлеров
func StripSpoilers(text string) string {
	if text == "" {
		return text
	}

	re := regexp.Prepare("stripSpoilers", `(?s)<tg-spoiler>(.+?)</tg-spoiler>`)
	text = re.ReplaceAll(text, "")

	return text
}

// Удаление пустых ссылок
func StripEmptyLinks(text string) string {
	if text == "" {
		return text
	}

	re := regexp.Prepare("stripEmptyLinks", `<a href="[^"]*?".*?></a>`)
	text = re.ReplaceAll(text, "")

	return text
}

// Удаление невидимых символов
func StripHidden(text string) string {
	if text == "" {
		return text
	}

	r := strings.NewReplacer(
		"\u00a0", "", "\u180e", "", "\u200b", "", "\u200c", "", "\u200d", "",
		"\u2060", "", "\u2061", "", "\u2062", "", "\u2063", "", "\ufeff", "",
	)
	text = r.Replace(text)

	return text
}

// Удаление невидимых символов (regexp)
func StripHiddenRegexp(text string) string {
	if text == "" {
		return text
	}

	re := regexp.Prepare("stripHidden",
		"[\u00a0\u180e\u200b\u200c\u200d"+
			"\u2060\u2061\u2062\u2063\ufeff]")

	text = re.ReplaceAll(text, "")

	return text
}

// Удаление мусора
func Clean(text string) string {
	if text == "" {
		return text
	}

	re := regexp.Prepare("slimNewlineTags", `((\r?\n\r?)+)((</[^>]+?>)+)`)
	text = re.ReplaceAll(text, "$3$1")

	re = regexp.Prepare("cleanEmoji", `<i class="emoji".*?><.+?>(.+?)</.+?></i>`)
	text = re.ReplaceAll(text, "$1")

	re = regexp.Prepare("cleanEmojiTg", `<tg-emoji.*?>(.+?)</tg-emoji>`)
	text = re.ReplaceAll(text, "$1")

	re = regexp.Prepare("cleanHashtags", `<a href="\?q=.+?".*?>(.+?)</a>`)
	text = re.ReplaceAll(text, "$1")

	re = regexp.Prepare("slimFatLinks", `(?s)<a href="([^"]+?)"[^>]+?>(.*?)</a>`)
	text = re.ReplaceAll(text, `<a href="$1">$2</a>`)

	return text
}

// Удаление пробелов по краям
func Trim(text string) string {
	if text == "" {
		return text
	}

	text = strings.TrimSpace(text)

	return text
}

// Формирование строковой даты
func Date(date time.Time) string {
	return date.Format("2006-01-02 15:04:05")
}

// Формирование ссылки
func Link(link string) string {
	if link == "" {
		return link
	}

	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		link = "http://" + link
	}

	return link
}

// Удаление лишнего у ссылки
func TrimLink(link string) string {
	if link == "" {
		return link
	}

	link = strings.TrimPrefix(link, "https://")
	link = strings.TrimPrefix(link, "http://")
	link = strings.TrimSuffix(link, "/")

	return link
}

// Обрезка длинной строки
func Truncate(str string, lim uint) string {
	if len(str) <= int(lim) {
		return str
	}

	return str[:lim] + "..."
}
