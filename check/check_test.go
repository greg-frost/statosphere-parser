package check

import (
	"reflect"
	"testing"

	"statosphere/parser/links"
)

func TestUsername(t *testing.T) {
	tests := []struct {
		test     string
		value    string
		isStrict bool
		username string
		joinchat string
		post     uint
	}{
		{"Username", "username", false, "username", "", 0},
		{"UsernameShort", "user", false, "", "", 0},
		{"UsernameNumber", "000user", false, "", "", 0},
		{"UsernameRussian", "юзернейм", false, "", "", 0},
		{"UsernameStrict", "username", true, "", "", 0},
		{"UsernameAt", "@username", true, "username", "", 0},
		{"UsernameMail", "username@mail.com", false, "username", "", 0},
		{"UsernameMailStrict", "username@mail.com", true, "", "", 0},
		{"UsernameTme", "t.me/username", true, "username", "", 0},
		{"UsernameTelegramMe", "telegram.me/username", true, "username", "", 0},
		{"UsernameHttp", "http://t.me/username", true, "username", "", 0},
		{"UsernameHttps", "https://t.me/username", true, "username", "", 0},
		{"UsernameDomain", "https://username.t.me", true, "username", "", 0},
		{"UsernameResolve", "tg://resolve?domain=username", true, "username", "", 0},
		{"UsernamePost", "@username/100", true, "username", "", 100},
		{"JoinchatPlus", "+abc4_fGhI0-LmnOp", false, "", "abc4_fGhI0-LmnOp", 0},
		{"JoinchatPrefix", "joinchat/abc4_fGhI0-LmnOp", false, "", "abc4_fGhI0-LmnOp", 0},
		{"JoinchatTme", "t.me/+abc4_fGhI0-LmnOp", false, "", "abc4_fGhI0-LmnOp", 0},
		{"JoinchatShort", "t.me/joinchat/abc4_fGhI0-Lm", false, "", "", 0},
		{"JoinchatPost", "+abc4_fGhI0-LmnOp/200th", false, "", "abc4_fGhI0-LmnOp", 200},
		{"JoinchatInvite", "tg://join?invite=abc4_fGhI0-LmnOp", false, "", "abc4_fGhI0-LmnOp", 0},
		{"JoinchatLongPost", "+AAAAAEabc4_fGhI0-LmnOp/300/ext", false, "", "AAAAAEabc4_fGhI0-LmnOp", 300},
		{"JoinchatChars", "joinchat/AAAAAEab3D*EfGh0+KLmnO", false, "", "", 0},
		{"Empty", "", false, "", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			username, joinchat, post := Username(tt.value, tt.isStrict)

			if username != tt.username {
				t.Errorf("Username - получено значение: %v, ожидается: %v", username, tt.username)
			}
			if joinchat != tt.joinchat {
				t.Errorf("Joinchat - получено значение: %v, ожидается: %v", joinchat, tt.joinchat)
			}
			if post != tt.post {
				t.Errorf("Post - получено значение: %v, ожидается: %v", post, tt.post)
			}
		})
	}
}

func TestLinks(t *testing.T) {
	none := links.New()

	one := links.New()
	one.Add("www.example.com", "", 0)

	caption := links.New()
	caption.Add("www.example.com", "Example", 0)

	two := links.New()
	two.Add("www.site.ru", "", 0)
	two.Add("www.site.com", "Site", 0)

	captions := links.New()
	captions.Add("site.io", "Rus", 0)
	captions.Add("site.io", "Eng", 0)

	query := links.New()
	query.Add("example.com/?param=value#anchor", "", 0)

	email := links.New()
	email.Add("username@mail.com", "", 0)

	usernames := links.New()
	usernames.Add("@username", "", 0)
	usernames.Add("t.me/username", "", 0)
	usernames.Add("https://t.me/username", "", 0)

	joinchat := links.New()
	joinchat.Add("t.me/+abc4_fGhI0-LmnOp", "", 0)

	userchat := links.New()
	userchat.Add("@username", "+abc4_fGhI0-LmnOp", 0)

	joinname := links.New()
	joinname.Add("+abc4_fGhI0-LmnOp", "@username", 0)

	posts := links.New()
	posts.Add("@username/100", "", 0)
	posts.Add("t.me/username/200", "", 0)

	tests := []struct {
		test     string
		value    string
		isStrict bool
		result   links.Links
	}{
		{"OneLink", "Text with www.example.com link", false, one},
		{"OneStrict", "Text with www.example.com link", true, none},
		{"OneTag", `Text with <a href="www.example.com"></a> link`, false, one},
		{"OneCaption", `Text with <a href="www.example.com">Example</a> link`, false, caption},
		{"OneSameCaption", `Text with <a href="www.example.com">www.example.com</a> link`, false, one},
		{"TwoLinks", `Text with www.site.ru and <a href="www.site.com">Site</a> links`, false, two},
		{"TwoCaptions", `Text with <a href="site.io">Rus</a> and <a href="site.io">Eng</a> links`, false, captions},
		{"Query", `Text with example.com/?param=value#anchor link`, false, query},
		{"Email", "Text with username@mail.com email-link", false, email},
		{"Usernames", "@username, t.me/username, https://t.me/username, username.t.me", true, usernames},
		{"Joinchat", "+abc4_fGhI0-LmnOp, joinchat/abc4_fGhI0-LmnOp, t.me/+abc4_fGhI0-LmnOp", false, joinchat},
		{"Userchat", `Text with <a href="@username">+abc4_fGhI0-LmnOp</a> link`, false, userchat},
		{"Joinname", `Text with <a href="+abc4_fGhI0-LmnOp">@username</a> link`, false, joinname},
		{"Posts", `Text with @username/100 and t.me/username/200 links`, false, posts},
		{"Empty", "", false, none},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Links(tt.value, tt.isStrict)

			var matches int
			for i, g := range result {
				for j, e := range tt.result {
					if i == j && e.Link == g.Link && e.Count == g.Count &&
						reflect.DeepEqual(e.Captions, g.Captions) &&
						reflect.DeepEqual(e.Posts, g.Posts) {
						matches++
					}
				}
			}

			if matches != len(result) || matches != len(tt.result) {
				t.Errorf("Получено значение: %#v, ожидается: %#v", result, tt.result)
			}
		})
	}
}

func TestAdvLinks(t *testing.T) {
	none := links.New()

	list := links.New()
	list.Add("@username1", "", 0)
	list.Add("t.me/username2/100", "Caption", 0)
	list.Add("https://t.me/username3", "t.me/joinchat/abc4_fGhI0-LmnOp/200", 0)
	list.Add("usernameN", "", 0)
	list.Add("www.example.com", "", 0)
	list.Add("https://site.com", "", 0)
	list.Add("site.ru", "", 0)

	usernames := links.New()
	usernames.Add("https://t.me/username1", "", 0)
	usernames.Add("https://t.me/username2", "", 100)
	usernames.Add("https://t.me/username3", "https://t.me/joinchat/abc4_fGhI0-LmnOp", 200)

	notSelf := links.New()
	notSelf.Add("https://t.me/username1", "", 0)
	notSelf.Add("https://t.me/username2", "", 100)

	notSiblings := links.New()
	notSiblings.Add("https://t.me/username3", "https://t.me/joinchat/abc4_fGhI0-LmnOp", 200)

	username := links.New()
	username.Add("https://t.me/username", "", 0)

	captions := links.New()
	captions.Add("@username", "https://t.me/username", 0)

	siblings := links.New()
	siblings.Add("t.me/joinchat/abc4_fGhI0-LmnOp", "https://t.me/username", 0)

	site := links.New()
	site.Add("www.example.com", "Example", 0)

	tests := []struct {
		test     string
		values   links.Links
		self     string
		siblings links.Links
		result   links.Links
	}{
		{"Username", list, "", none, usernames},
		{"NotSelf", list, "https://t.me/username3", none, notSelf},
		{"NotSelfBad", list, "@username", none, usernames},
		{"NotSelfCaption", list, "https://t.me/joinchat/abc4_fGhI0-LmnOp", none, notSelf},
		{"NotSibling", list, "", notSelf, notSiblings},
		{"NotCaption", captions, "", none, username},
		{"NotSiblingCaption", siblings, "", username, none},
		{"NotUsername", site, "", none, none},
		{"Empty", none, "", none, none},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := AdvLinks(tt.values, tt.self, tt.siblings)

			var matches int
			for i, g := range result {
				for j, e := range tt.result {
					if i == j && e.Link == g.Link && e.Count == g.Count &&
						reflect.DeepEqual(e.Captions, g.Captions) &&
						reflect.DeepEqual(e.Posts, g.Posts) {
						matches++
					}
				}
			}

			if matches != len(result) || matches != len(tt.result) {
				t.Errorf("Получено значение: %#v, ожидается: %#v", result, tt.result)
			}
		})
	}
}

func TestNumber(t *testing.T) {
	tests := []struct {
		test    string
		value   string
		result  int
		isError bool
	}{
		{"Valid", " 500 ", 500, false},
		{"Negative", " -1 000k ", -1000000, false},
		{"FloatDot", "1.2K", 1200, false},
		{"FloatHyphen", "5,225M", 5225000, false},
		{"Text", "Ten", 0, true},
		{"TextBefore", "Twenty.350K", 350, false},
		{"TextAfter", ".30Mth", 300000, false},
		{"Empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, err := Number(tt.value)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPositiveNumber(t *testing.T) {
	tests := []struct {
		test    string
		value   string
		result  uint
		isError bool
	}{
		{"Valid", " 500 ", 500, false},
		{"Negative", " -1 000k ", 0, false},
		{"FloatDot", "1.2K", 1200, false},
		{"FloatHyphen", "-5,225M", 0, false},
		{"Text", "Ten", 0, true},
		{"TextBefore", "Twenty--.350K", 0, true},
		{"TextAfter", ".30Mth", 300000, false},
		{"Empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, err := PositiveNumber(tt.value)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		test    string
		value   string
		result  int
		isError bool
	}{
		{"Valid", " 500 ", 500, false},
		{"Negative", " -1 000 ", -1000, false},
		{"FloatDot", "1.0", 1, false},
		{"FloatHyphen", "5,0", 5, false},
		{"Text", "Ten", 0, true},
		{"TextBefore", "Twenty20", 20, false},
		{"TextAfter", "30th", 30, false},
		{"Empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, err := Int(tt.value)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPositiveInt(t *testing.T) {
	tests := []struct {
		test    string
		value   string
		result  uint
		isError bool
	}{
		{"Valid", " 500 ", 500, false},
		{"Negative", " -1 000 ", 0, false},
		{"FloatDot", "1.0", 1, false},
		{"FloatHyphen", "-5,0", 0, false},
		{"Text", "Ten", 0, true},
		{"TextBefore", "Twenty20", 20, false},
		{"TextAfter", "-30th", 0, false},
		{"Empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, err := PositiveInt(tt.value)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		test    string
		value   string
		result  bool
		isError bool
	}{
		{"True", "True", true, false},
		{"Yes", "yes", true, false},
		{"On", "ON", true, false},
		{"Off", "OFF", false, false},
		{"Ok", "ok", false, false},
		{"One", "1", true, false},
		{"Zero", "0", false, false},
		{"Enabled", "EnableD", false, false},
		{"Disabled", "dISABLEd", false, false},
		{"Empty", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, err := Bool(tt.value)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		test    string
		value   string
		results []string
		isError bool
	}{
		{"OneItem", "{username}", []string{"username"}, false},
		{"TwoItems", " [ user1,user2]", []string{"user1", "user2"}, false},
		{"TwoEmpty", "(,)", []string{"", ""}, false},
		{"ThreeItems", "user1, user2, user3", []string{"user1", "user2", "user3"}, false},
		{"BadBrackets", "/username/", []string{"/username/"}, false},
		{"Empty", "   ", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			results, err := List(tt.value)

			if (err != nil) != tt.isError {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", err, tt.isError)
			}
			if !reflect.DeepEqual(results, tt.results) {
				t.Errorf("Получены значения: %v, ожидается: %v", results, tt.results)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		test   string
		value  int
		min    int
		result int
	}{
		{"EqualMin", 0, 0, 0},
		{"LowerThanMin", -1, 0, 0},
		{"BiggerThanMin", 2, 0, 2},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Min(tt.value, tt.min)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		test   string
		value  int
		max    int
		result int
	}{
		{"EqualMax", 0, 0, 0},
		{"LowerThanMax", -1, 0, -1},
		{"BiggerThanMax", 2, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Max(tt.value, tt.max)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestBetween(t *testing.T) {
	tests := []struct {
		test   string
		value  int
		min    int
		max    int
		result int
	}{
		{"Between", 5, 0, 0, 0},
		{"Range", 5, 0, 10, 5},
		{"SwapRange", 5, 10, 0, 5},
		{"LowerThanMin", -5, 0, 10, 0},
		{"BiggerThanMax", 15, 0, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Between(tt.value, tt.min, tt.max)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}
