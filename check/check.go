package check

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"statosphere/parser/format"
	"statosphere/parser/links"
	"statosphere/parser/regexp"
)

// Общие паттерны для проверок
const (
	PatternA          = `A-Za-z`
	PatternAN         = `A-Za-z0-9`
	PatternARN        = `A-Za-zА-Яа-яЁё0-9`
	PatternHttp       = `https?://`
	PatternSite       = `([` + PatternAN + `_-]+\.)?[` + PatternAN + `_-]+\.[` + PatternA + `][` + PatternAN + `]+`
	PatternQuery      = `[/?#][` + PatternARN + `/?#&%;:=+_.-]+[` + PatternARN + `/]`
	PatternEmail      = `(mailto://)?[` + PatternEmails + `]+@`
	PatternEmails     = PatternAN + `_\.\-`
	PatternTg         = `t\.me`
	PatternTgSite     = `(t|telegram)\.me/(s/)?`
	PatternTgResolve  = `tg://resolve\?domain=`
	PatternTgInvite   = `tg://join\?invite=`
	PatternUsername   = `[` + PatternA + `][` + PatternAN + `_]{4,31}`
	PatternJoinPrefix = `(joinchat/|\+)`
	PatternJoinHash   = `[` + PatternAN + `_-]{16,22}`
	PatternPost       = `\d+`
)

// Разбор юзернейма (джойнчата) канала
func Username(value string, isStrict bool) (username, joinchat string, post uint) {
	if value == "" {
		return
	}

	var strictCond string
	if !isStrict {
		strictCond = "?"
	}

	pattern := `(?i)((^|[^` + PatternEmails + `/])@(?P<usernameAt>` + PatternUsername + `)|` + PatternTgResolve +
		`(?P<usernameRs>` + PatternUsername + `)|` + PatternTgInvite + `(?P<joinchatIv>` + PatternJoinHash + `)` +
		`|(` + PatternHttp + `)?(?P<usernameDm>` + PatternUsername + `)\.` + PatternTg + `|(((` + PatternHttp + `)?` +
		PatternTgSite + `)` + strictCond + `(` + PatternJoinPrefix + `(?P<joinchat>` + PatternJoinHash + `)|` +
		`(?P<username>` + PatternUsername + `))))(/(?P<post>` + PatternPost + `))?`

	re := regexp.Prepare("username", pattern)
	res := re.Find(value)

	if len(res) > 0 {
		switch {
		case res["usernameAt"] != "":
			username = res["usernameAt"]
		case res["usernameRs"] != "":
			username = res["usernameRs"]
		case res["usernameDm"] != "":
			username = res["usernameDm"]
		case res["username"] != "":
			username = res["username"]
		}
		if username == "joinchat" || username == "addstickers" {
			username = ""
		}

		switch {
		case res["joinchatIv"] != "":
			joinchat = res["joinchatIv"]
		case res["joinchat"] != "":
			joinchat = res["joinchat"]
		}

		post, _ = PositiveInt(res["post"])
	}

	return
}

// Разбор ссылок html-страницы
func Links(value string, isStrict bool) links.Links {
	links := links.New()

	if value == "" {
		return links
	}

	var strictCond string
	if !isStrict {
		strictCond = "?"
	}

	pattern := `(?is)(<a href="(?P<url>[^"]+?)"[^>]*?>(?P<caption>.*?)</a>|(^|[^` + PatternEmails + `])` +
		`(?P<link>(` + PatternEmail + PatternSite + `|(` + PatternHttp + `)` + strictCond + PatternSite +
		`(` + PatternQuery + `)?|@` + PatternUsername + `(/` + PatternPost + `)?|((` + PatternHttp + `)?` +
		PatternTgSite + `(` + PatternJoinPrefix + PatternJoinHash + `|` + PatternUsername + `)` +
		`(/` + PatternPost + `)?))))`

	re := regexp.Prepare("links", pattern)
	ress := re.FindAll(value)

	if len(ress) > 0 {
		for _, res := range ress {
			var link, caption string
			switch {
			case res["url"] != "":
				link = res["url"]
				caption = res["caption"]
			case res["link"] != "":
				link = res["link"]
			}

			links.Add(link, caption, 0)
		}
	}

	return links
}

// Разбор рекламных ссылок (на другие каналы)
func AdvLinks(values links.Links, self string, siblings links.Links) links.Links {
	advs := links.New()

	if len(values) == 0 {
		return advs
	}

	for _, data := range values {

		var (
			username, joinchat string
			caption            string
			post, cPost        uint
		)

		// Ссылка

		username, joinchat, post = Username(data.Link, true)
		_, link := format.Username(username, joinchat, 0)

		if link == "" {
			continue
		}
		if link == self {
			continue
		}
		if siblings.IsExist(link) {
			continue
		}

		// Описание

		for _, linkCaption := range data.Captions {
			username, joinchat, cPost = Username(linkCaption, true)
			_, caption = format.Username(username, joinchat, 0)
			if caption != "" {
				break
			}
		}

		if caption != "" {
			if caption == self {
				continue
			}
			if siblings.IsExist(caption) {
				continue
			}
			if caption == link {
				caption = ""
			}
		}

		if post == 0 && cPost != 0 {
			post = cPost
		}

		advs.Add(link, caption, post)
	}

	return advs
}

// Проверка и получение числа (с коэффициентами)
func Number(value string) (int, error) {
	if value == "" {
		return 0, errors.New("строка пуста")
	}

	var (
		numbers string
		koef    float64 = 1
	)

	for _, char := range value {
		switch {
		case char >= '0' && char <= '9' || char == '-' || char == '.' || char == ',':
			numbers += string(char)
		case char == 'K' || char == 'k':
			koef = 1000
		case char == 'M' || char == 'm':
			koef = 1000000
		}
	}

	numbers = strings.ReplaceAll(numbers, ",", ".")
	floatVal, err := strconv.ParseFloat(numbers, 64)
	if err != nil {
		return 0, fmt.Errorf("строка %q не переведена в тип float", numbers)
	}
	intVal := int(math.Round(floatVal * koef))

	return intVal, nil
}

// Проверка и получение положительного числа (с коэффициентами)
func PositiveNumber(value string) (uint, error) {
	val, err := Number(value)
	val = Min(val, 0)

	return uint(val), err
}

// Проверка и получение числа
func Int(value string) (int, error) {
	if value == "" {
		return 0, errors.New("строка пуста")
	}

	var numbers string

outer:
	for _, char := range value {
		switch {
		case char >= '0' && char <= '9' || char == '-':
			numbers += string(char)
		case char == '.' || char == ',':
			break outer
		}
	}

	intVal, err := strconv.Atoi(numbers)
	if err != nil {
		return 0, fmt.Errorf("строка %q не переведена в тип int", value)
	}

	return intVal, nil
}

// Проверка и получение положительного числа
func PositiveInt(value string) (uint, error) {
	val, err := Int(value)
	val = Min(val, 0)

	return uint(val), err
}

// Проверка логического значения
func Bool(value string) (bool, error) {
	if value == "" {
		return false, errors.New("строка пуста")
	}

	value = strings.ToLower(value)

	switch value {
	case "true", "yes", "on", "1":
		return true, nil
	default:
		return false, nil
	}
}

// Проверка и получение списка значений
func List(value string) ([]string, error) {
	list := make([]string, 0)

	value = strings.ReplaceAll(value, " ", "")
	value = strings.Trim(value, "()[]{}")

	if value == "" {
		return list, errors.New("строка пуста")
	}

	list = strings.Split(value, ",")

	return list, nil
}

// Проверка на нижний предел
func Min(value, min int) int {
	if value < min {
		value = min
	}

	return value
}

// Проверка на верхний предел
func Max(value, max int) int {
	if value > max {
		value = max
	}

	return value
}

// Проверка на диапазон
func Between(value, min, max int) int {
	if min > max {
		min, max = max, min
	}

	value = Min(value, min)
	value = Max(value, max)

	return value
}
