package format

import (
	"testing"
	"time"
)

func TestUsername(t *testing.T) {
	tests := []struct {
		test     string
		username string
		joinchat string
		post     uint
		peer     string
		link     string
	}{
		{"Username", "username", "", 0, "@username", "https://t.me/username"},
		{"UsernamePost", "username", "", 1, "@username/1", "https://t.me/username/1"},
		{"UsernameJoinchat", "username", "abc4_fGhI0-LmnOp", 0, "@username", "https://t.me/username"},
		{"Joinchat", "", "abc4_fGhI0-LmnOp", 0, "+abc4_fGhI0-LmnOp", "https://t.me/joinchat/abc4_fGhI0-LmnOp"},
		{"JoinchatPost", "", "abc4_fGhI0-LmnOp", 2, "+abc4_fGhI0-LmnOp/2", "https://t.me/joinchat/abc4_fGhI0-LmnOp/2"},
		{"Empty", "", "", 0, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			peer, link := Username(tt.username, tt.joinchat, tt.post)

			if peer != tt.peer {
				t.Errorf("Peer - получено значение: %v, ожидается: %v", peer, tt.peer)
			}
			if link != tt.link {
				t.Errorf("Link - получено значение: %v, ожидается: %v", link, tt.link)
			}
		})
	}
}

func TestPageLink(t *testing.T) {
	tests := []struct {
		test     string
		username string
		joinchat string
		info     string
		messages string
	}{
		{"Username", "username", "", "https://t.me/username", "https://t.me/s/username"},
		{"UsernameJoinchat", "username", "abc4_fGhI0-LmnOp", "https://t.me/username", "https://t.me/s/username"},
		{"Joinchat", "", "abc4_fGhI0-LmnOp", "https://t.me/joinchat/abc4_fGhI0-LmnOp", ""},
		{"Empty", "", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			info, messages := PageLink(tt.username, tt.joinchat)

			if info != tt.info {
				t.Errorf("Info - получено значение: %v, ожидается: %v", info, tt.info)
			}
			if messages != tt.messages {
				t.Errorf("Messages - получено значение: %v, ожидается: %v", messages, tt.messages)
			}
		})
	}
}

func TestStripPage(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Head", "<doctype html><head>...</head><header>\n</header><body></body>", "<doctype html><body></body>"},
		{"Svg", "Text<svg>...</svg>able <svg>\n...\n</svg>text ends<svg attr>no</svg>", "Textable text ends"},
		{"Tgme", "<div class=\"tgme_widget_message_user\">\n</div>\"></div>", `<div class=""></div>`},
		{"NoMatch", "<doctype html><html><body></body></html>", "<doctype html><html><body></body></html>"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := StripPage(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestStripPageRegexp(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Head", "<doctype html><head>...</head><header>\n</header><body></body>", "<doctype html><body></body>"},
		{"Svg", "Text<svg>...</svg>able <svg>\n...\n</svg>text ends<svg attr>no</svg>", "Textable text ends"},
		{"Tgme", "<div class=\"tgme_widget_message_user\">\n</div>\"></div>", `<div class=""></div>`},
		{"NoMatch", "<doctype html><html><body></body></html>", "<doctype html><html><body></body></html>"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := StripPageRegexp(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestStripSegments(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		open   string
		close  string
		result string
	}{
		{"Small", "<tag>...</tag><b>Text</b>", "<tag>", "</tag>", "<b>Text</b>"},
		{"Big", "<tag>\n</tag><b>Text</b>NoText", "<tag>", "</b>", "NoText"},
		{"OpenMissing", "...</tag>", "<tag>", "</tag>", "...</tag>"},
		{"CloseMissing", "<tag>\n", "<tag>", "</tag>", "<tag>\n"},
		{"Swap", "<tag>...</tag>", "</tag>", "<tag>", "<tag>...</tag>"},
		{"Empty", "", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := StripSegments(tt.value, tt.open, tt.close)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestSafeHtml(t *testing.T) {
	var (
		value  = "  This is a\u00a0 simple <b>text<br></b> that <a href=\"link\" class=\"me\">I&#39;m</a> wrote\t  "
		result = "This is a simple <b>text</b>\n that <a href=\"link\">I'm</a> wrote"
	)

	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Valid", value, result},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := SafeHtml(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestSafeString(t *testing.T) {
	var (
		value  = "  This is a\u00a0 simple <b>text<br></b> that <a href=\"link\" class=\"me\">I&#39;m</a> wrote\t  "
		result = "This is a simple text that I'm wrote"
	)

	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Valid", value, result},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := SafeString(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestSafeText(t *testing.T) {
	var (
		value  = "  This is a\u00a0 simple <b>text<br></b> that <a href=\"link\" class=\"me\">I&#39;m</a> wrote\t  "
		result = "This is a simple text\n that I'm wrote"
	)

	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Valid", value, result},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := SafeText(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestUnescape(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Quote", "&quot;Quote&quot;", `"Quote"`},
		{"Ampersand", `&amp;param=amp;`, "&param=amp;"},
		{"Apostrophe", "I&#39;m &39;", "I'm &39;"},
		{"Tags", `&lt;tag&gt;`, "<tag>"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Unescape(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestNewlines(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Break", `L<br>I<br />N<br clear="both">E`, "L\nI\nN\nE"},
		{"Tab", "T\t\tA\t\tB", "T\n\nA\n\nB"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Newlines(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestStripNewlines(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Break", `L<br>I<br />N<br clear="both">E`, "LINE"},
		{"Break", "\nLinux\r\nWindows\n\rMacOS", "LinuxWindowsMacOS"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := StripNewlines(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestStripTags(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Tag", "T<b>E</b>X<i>T</i>", "TEXT"},
		{"Attr", `T<b class="bold">EX<i>T</i></b>`, "TEXT"},
		{"Inner", "T<b>EX<i>T</i></b>", "TEXT"},
		{"Swap", "T<b>EX<i></b>T</i>", "TEXT"},
		{"Bad", "T<bad />EX<i>T</b>", "TEXT"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := StripTags(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestStripServiceTags(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Code", "Not<code>alert('vzlom')\nadmins++</code>Code", "NotCode"},
		{"Pre", "Not<pre>\n.\n.\n.</pre>Pre", "NotPre"},
		{"Spoiler", "Not<tg-spoiler>shut\nup!</tg-spoiler>Spoiler", "NotSpoiler"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := StripServiceTags(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestStripCode(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Code", "Not<code>alert('vzlom')\nadmins++</code>Code", "NotCode"},
		{"Pre", "Not<pre>\n.\n.\n.</pre>Pre", "NotPre"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := StripCode(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestStripSpoilers(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Spoiler", "Not<tg-spoiler>shut\nup!</tg-spoiler>Spoiler", "NotSpoiler"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := StripSpoilers(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestStripEmptyLinks(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Caption", `<a href="https://example.com">Example</a>`, `<a href="https://example.com">Example</a>`},
		{"NoCaption", `<a href="https://example.com"></a>`, ""},
		{"Attr", `<a href="" onclick="alert('test')"></a>`, ""},
		{"NoLink", `<a href="">Example</a>`, `<a href="">Example</a>`},
		{"Nothing", `<a href=""></a>`, ""},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := StripEmptyLinks(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestStripHidden(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Hidden", "\u200bH\u00a0I\u2062\u180eD\u2063\u2060D\ufeffE\u2061\u200dN\u200c", "HIDDEN"},
		{"NoHidden", "\u2055HIDDEN\u2075", "\u2055HIDDEN\u2075"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := StripHidden(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestStripHiddenRegexp(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Hidden", "\u200bH\u00a0I\u2062\u180eD\u2063\u2060D\ufeffE\u2061\u200dN\u200c", "HIDDEN"},
		{"NoHidden", "\u2055HIDDEN\u2075", "\u2055HIDDEN\u2075"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := StripHiddenRegexp(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"SlimNewlines", "<b>B\n</b>\n<u><i>UI\r\n\n\r</i></u>", "<b>B</b>\n\n<u><i>UI</i></u>\r\n\n\r"},
		{"Emoji", `<i class="emoji" attr="value"><b class="bold">Emoji</b></i>`, "Emoji"},
		{"EmojiTg", `<tg-emoji attr="value">Emoji</tg-emoji>`, "Emoji"},
		{"Hashtag", `<a href="?q=%23%D1%80%D0%B0..." attr="value">#Hashtag</a>`, "#Hashtag"},
		{"FatLink", `<a href="link" onclick="script">Caption</a>`, `<a href="link">Caption</a>`},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Clean(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestTrim(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"LTrim", "   Text", "Text"},
		{"RTrim", "Text   ", "Text"},
		{"Trim", "   Text  ", "Text"},
		{"NoTrim", "T e x t", "T e x t"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Trim(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %q, ожидается: %q", result, tt.result)
			}
		})
	}
}

func TestDate(t *testing.T) {
	var (
		dateString  = "2006-01-02 15:04:05"
		dateTime, _ = time.Parse("2006-01-02 15:04:05", dateString)
	)

	tests := []struct {
		test   string
		date   time.Time
		result string
	}{
		{"Valid", dateTime, dateString},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Date(tt.date)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestLink(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Http", "http://example.com", "http://example.com"},
		{"Https", "https://example.com", "https://example.com"},
		{"None", "example.com", "http://example.com"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Link(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestTrimLink(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result string
	}{
		{"Http", "http://example.com", "example.com"},
		{"Https", "https://example.com", "example.com"},
		{"EndSlash", "https://example.com/", "example.com"},
		{"NotEndSlash", "https://example.com/page", "example.com/page"},
		{"None", "example.com", "example.com"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := TrimLink(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		lim    uint
		result string
	}{
		{"Short", "Text", 5, "Text"},
		{"Long", "Truncated text", 9, "Truncated..."},
		{"Russian", "Текст", 6, "Тек..."},
		{"Bad", "Текст", 5, "Те\xd0..."},
		{"Zero", "Text", 0, "..."},
		{"Empty", "", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := Truncate(tt.value, tt.lim)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}
