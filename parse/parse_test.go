package parse

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"statosphere/parser/cache"
	"statosphere/parser/channel"
	"statosphere/parser/channels"
	"statosphere/parser/get"
	"statosphere/parser/links"
	"statosphere/parser/message"
	"statosphere/parser/mock"
	"statosphere/parser/response"
	"statosphere/parser/value"
)

func init() {
	os.Chdir("..")
	get.SetTransport("file", 0)
}

func TestNewChannels(t *testing.T) {
	tests := []struct {
		test   string
		result channels.Channels
	}{
		{"Valid", channels.New()},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := NewChannels().Channels

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestParse(t *testing.T) {
	valid := NewChannels()

	valid.Add("thecodemedia")
	valid.Add("glav_hack")
	valid.Add("not_existed_channel")

	empty := NewChannels()

	ctx := context.Background()
	timeout, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	tests := []struct {
		test     string
		value    *Channels
		ctx      context.Context
		isExact  bool
		messages uint
		isCache  bool
		result   int
		errs     []error
	}{
		{"Exact", &valid, ctx, true, 0, false, 2, []error{
			errors.New("Ошибка парсинга информации: нет данных для @not_existed_channel"),
		}},
		{"NotExact", &valid, ctx, false, 0, true, 1, []error{
			errors.New("Ошибка парсинга сообщений: нет данных для @glav_hack"),
			errors.New("Ошибка парсинга сообщений: нет данных для @not_existed_channel"),
		}},
		{"Cache", &valid, ctx, true, 0, true, 2, []error{
			errors.New("Ошибка парсинга информации: нет данных для @not_existed_channel"),
		}},
		{"Messages", &valid, ctx, true, 5, true, 2, []error{
			errors.New("Ошибка парсинга информации: нет данных для @not_existed_channel"),
			errors.New("Ошибка парсинга сообщений: нет данных для @not_existed_channel"),
			errors.New("Ошибка парсинга сообщений: нет данных для @glav_hack"),
		}},
		{"Cancel", &valid, timeout, true, 5, false, 0, []error{
			errors.New("Ошибка парсинга: отмена для @thecodemedia"),
			errors.New("Ошибка парсинга: отмена для @glav_hack"),
			errors.New("Ошибка парсинга: отмена для @not_existed_channel"),
		}},
		{"Empty", &empty, ctx, false, 0, false, 0, []error{errors.New("Ошибка парсинга: нет каналов")}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			cache.Disable()
			if tt.isCache {
				cache.Enable()
			}

			result, errs := tt.value.Parse(tt.ctx, tt.isExact, tt.messages)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}

			var matches int
			for _, gErr := range tt.errs {
				for _, eErr := range tt.errs {
					if fmt.Sprint(gErr) == fmt.Sprint(eErr) {
						matches++
					}
				}
			}

			if matches != len(errs) || matches != len(tt.errs) {
				t.Errorf("Получены ошибки: %q, ожидаются: %q", errs, tt.errs)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	tests := []struct {
		test         string
		value        string
		result       bool
		code         int
		isError      bool
		isPanic      bool
		peer         string
		link         string
		title        string
		hasAbout     bool
		contacts     []string
		siblings     []string
		hasImage     bool
		kind         string
		participants value.Value
		isVerified   bool
		isScam       bool
	}{
		{"Username", "codecamp", true, 200, false, false, "@codecamp",
			"https://t.me/codecamp", "CodeCamp", true, []string{"@camprobot"},
			[]string{"t.me/camprobot", "t.me/workcamp"}, true, "channel",
			value.Value{Exact: 82932}, false, false},
		{"Joinchat", "+so8YUpEsL4BkZGQy", true, 200, false, false, "+so8YUpEsL4BkZGQy",
			"https://t.me/joinchat/so8YUpEsL4BkZGQy", "Джейпег Малевича", true,
			[]string{"t.me/dssale/264"}, []string{"https://t.me/Alivian"}, true, "private",
			value.Value{Exact: 189490}, false, false},
		{"Verified", "thecodemedia", true, 200, false, false, "@thecodemedia",
			"https://t.me/thecodemedia", "Журнал «Код»", true, []string{"thecode.media"}, []string{},
			true, "channel", value.Value{Exact: 63336}, true, false},
		{"Scam", "lolz_guru", true, 200, false, false, "@lolz_guru", "https://t.me/lolz_guru",
			"LOLZTEAM", false, []string{}, []string{}, true, "channel",
			value.Value{Exact: 163840}, false, true},
		{"Chat", "ru_python", true, 200, false, false, "@ru_python", "https://t.me/ru_python",
			"Python", true, []string{"t.me/ru_python/1961404"}, []string{},
			true, "chat", value.Value{Exact: 14035}, false, false},
		{"User", "username9", true, 200, false, false, "@username9",
			"https://t.me/username9", "User", false, []string{}, []string{}, true,
			"user", value.Value{}, false, false},
		{"Bot", "botfather", true, 200, false, false, "@botfather",
			"https://t.me/botfather", "BotFather", true, []string{}, []string{}, true,
			"bot", value.Value{}, true, false},
		{"NotFound", "not_existed_channel", false, 404, true, false, "@not_existed_channel",
			"https://t.me/not_existed_channel", "", false, []string{}, []string{},
			false, "", value.Value{}, false, false},
		{"NotValid", "not_valid_channel", false, 404, true, false, "@not_valid_channel",
			"https://t.me/not_valid_channel", "", false, []string{}, []string{},
			false, "", value.Value{}, false, false},
		{"NoTitle", "seniorpy", false, 500, false, true, "@seniorpy",
			"https://t.me/seniorpy", "", false, []string{}, []string{},
			false, "", value.Value{}, false, false},
		{"WrongParticipants", "netstalkers", false, 500, false, true, "@netstalkers",
			"https://t.me/netstalkers", "", false, []string{}, []string{},
			false, "", value.Value{}, false, false},
		{"NoParticipants", "vtosters", false, 500, false, true, "@vtosters",
			"https://t.me/vtosters", "", false, []string{}, []string{},
			false, "", value.Value{}, false, false},
		{"Empty", "", false, 500, true, false, "", "", "", false, []string{},
			[]string{}, false, "", value.Value{}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			var res response.Response
			var ch = make(chan response.Response)
			var wg sync.WaitGroup

			cc := NewChannels()
			cc.Add(tt.value)

			c, ok := cc.Find(tt.value)
			if !ok {
				c = &channel.Channel{Username: tt.value}
			}

			go func() {
				wg.Add(1)
				defer func() {
					if isPanic := recover() != nil; isPanic != tt.isPanic {
						t.Errorf("Получена паника: %v, ожидается: %v", isPanic, tt.isPanic)
					}
					wg.Done()
				}()
				Info(ch, *c)
			}()

			res = <-ch
			if res.Ok {
				*c = res.Data.(channel.Channel)
			}

			wg.Wait()

			if res.Ok != tt.result {
				t.Fatalf("Получен результат: %v, ожидается: %v", res.Ok, tt.result)
			}
			if res.Code != tt.code {
				t.Fatalf("Code - получено значение: %v, ожидается: %v", res.Code, tt.code)
			}
			if (res.Error != nil) != tt.isError && !tt.isPanic {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", res.Error, tt.isError)
			}

			if c.Peer != tt.peer {
				t.Errorf("Peer - получено значение: %v, ожидается: %v", c.Peer, tt.peer)
			}
			if c.Link != tt.link {
				t.Errorf("Link - получено значение: %v, ожидается: %v", c.Link, tt.link)
			}

			if !strings.Contains(c.Title, tt.title) {
				t.Errorf("Title - получено значение: %v, ожидается совпадение: %v", c.Title, tt.title)
			}
			if hasAbout := c.About != ""; hasAbout != tt.hasAbout {
				t.Errorf("About - получено значение: %v, ожидается: %v", hasAbout, tt.hasAbout)
			}

			for _, link := range tt.contacts {
				if !c.Contacts.IsExist(link) {
					t.Errorf("Contacts - получены значения: %#v, ожидаются совпадения: %q, не найдено: %q",
						c.Contacts, tt.contacts, link)
				}
			}
			for _, link := range tt.siblings {
				if !c.Siblings.IsExist(link) {
					t.Errorf("Siblings - получены значения: %#v, ожидаются совпадения: %q, не найдено: %q",
						c.Siblings, tt.siblings, link)
				}
			}

			if hasImage := c.Image != ""; hasImage != tt.hasImage {
				t.Errorf("Image - получено значение: %v, ожидается: %v", hasImage, tt.hasImage)
			}

			if c.Kind != tt.kind {
				t.Errorf("Kind - получено значение: %v, ожидается: %v", c.Kind, tt.kind)
			}

			if !reflect.DeepEqual(c.Participants, tt.participants) {
				t.Errorf("Participants - получено значение: %#v, ожидается: %#v", c.Participants, tt.participants)
			}

			if c.IsVerified != tt.isVerified {
				t.Errorf("IsVerified - получено значение: %v, ожидается: %v", c.IsVerified, tt.isVerified)
			}
			if c.IsScam != tt.isScam {
				t.Errorf("IsScam - получено значение: %v, ожидается: %v", c.IsScam, tt.isScam)
			}
		})
	}
}

func TestMessages(t *testing.T) {
	tests := []struct {
		test          string
		value         string
		result        bool
		code          int
		isError       bool
		isPanic       bool
		isParseInfo   bool
		title         string
		hasAbout      bool
		hasImage      bool
		participants  value.Value
		isVerified    bool
		isScam        bool
		messagesCount uint
		afterMsgId    uint
		isSkipEmpty   bool
		messages      map[uint]message.Message
	}{
		{"Username", "codecamp", true, 200, false, false, true, "CodeCamp",
			true, true, value.Value{Approx: 82900, Short: "82.9K"}, false, false,
			10, 2300, false, map[uint]message.Message{
				2375: {
					MessageHtml: "Тут недавно прошел Всемирный день паролей",
					Views:       value.Value{Approx: 4900, Short: "4.9K"},
					Date:        time.Date(2023, 5, 5, 8, 40, 8, 0, time.UTC),
					DateLocal:   time.Date(2023, 5, 5, 11, 40, 8, 0, time.Local),
				},
				2370: {
					MessageHtml: "Пора создавать общество защиты прав роботов",
					IsForwarded: true,
					FwdTitle:    "Киллер-фича",
					HasVideo:    true,
				}}},
		{"Joinchat", "+so8YUpEsL4BkZGQy", false, 404, true, false, true, "",
			false, false, value.Value{}, false, false,
			10, 0, false, nil},
		{"Verified", "thecodemedia", true, 200, false, false, true, "Журнал «Код»",
			true, true, value.Value{Approx: 63300, Short: "63.3K"}, true, false,
			30, 0, true, map[uint]message.Message{
				7140: {
					MessageHtml: "#подборка_Код",
					HasImage:    true,
					Attachments: []string{"EOJbXgx", "Y9f7p0i", "ck87ZM2"},
					Hashtags:    []string{"подборка_Код"},
				}}},
		{"Scam", "lolz_guru", true, 200, false, false, true, "LOLZTEAM",
			false, true, value.Value{Approx: 164000, Short: "164K"}, false, true,
			20, 0, false, map[uint]message.Message{
				3368: {
					MessageHtml: "CRAZY EVIL",
					HasImage:    true,
					Buttons: links.Links{
						"#1": {Link: "http://t.me/CrazyEvilNft_bot"},
					},
					Links: links.Links{
						"#1": {Link: "http://t.me/CrazyEvilNft_bot"},
						"#2": {Link: "https://zelenka.guru/threads/3454432/"},
					},
					Advs: links.Links{
						"#1": {Link: "http://t.me/CrazyEvilNft_bot"},
					},
					Hashtags: []string{"спонсор_месяца"},
				}}},
		{"Messages", "antichristone", true, 200, false, false, true, "Antichrist Blog",
			true, true, value.Value{Approx: 78600, Short: "78.6K"}, false, false,
			10, 0, false, map[uint]message.Message{
				500: {
					MessageHtml: "Какой вариант выбираете?",
					IsPoll:      true,
				},
				503: {
					MessageHtml: "t.me/antichristone_stream",
					IsEdited:    true,
					HasDocument: true,
				},
				521: {
					MessageHtml: "<b>Комп для чайника</b>",
					HasVideo:    true,
				}}},
		{"MessagesMedia", "procode404", true, 200, false, false, true, "[404]",
			true, true, value.Value{Approx: 62100, Short: "62.1K"}, false, false,
			10, 0, false, map[uint]message.Message{
				2332: {
					MessageHtml: "<a href=\"https://www.youtube.com/watch?v=GQfC0nYrto8\">" +
						"Перейти к просмотру</a>\n\n",
					HasVideo: true,
					Links: links.Links{
						"#1": {Link: "https://www.youtube.com/watch?v=GQfC0nYrto8"},
					},
					Hashtags: []string{"теория"},
				}}},
		{"MessagesDocs", "excelhackru", true, 200, false, false, true, "ЭКСЕЛЬ ХАК",
			true, true, value.Value{Approx: 53000, Short: "53K"}, true, false,
			10, 0, false, map[uint]message.Message{
				1350: {
					HasImage:    true,
					HasDocument: true,
					Attachments: []string{"Image.png (3.61 KB)"},
					Views:       value.Value{Approx: 19800, Short: "19.8K"},
				}}},
		{"MessagesNoId", "prg_memes", false, 500, false, true, true, "",
			false, false, value.Value{}, false, false,
			10, 0, false, nil},
		{"MessagesNoViews", "d_code", false, 500, false, true, true, "",
			false, false, value.Value{}, false, false,
			10, 0, false, nil},
		{"MessagesNoDate", "habr_com", false, 500, false, true, true, "",
			false, false, value.Value{}, false, false,
			10, 0, false, nil},
		{"NotFound", "not_existed_channel", false, 404, true, false, true, "",
			false, false, value.Value{}, false, false,
			10, 0, false, nil},
		{"NotValid", "not_valid_channel", false, 404, true, false, true, "",
			false, false, value.Value{}, false, false,
			10, 0, false, nil},
		{"NoTitle", "seniorpy", false, 500, false, true, true, "",
			false, false, value.Value{}, false, false,
			10, 0, false, nil},
		{"NoParse", "seniorpy", true, 200, false, false, false, "",
			false, false, value.Value{}, false, false,
			10, 0, false, nil},
		{"NoParticipants", "vtosters", false, 500, false, true, true, "",
			false, false, value.Value{}, false, false,
			10, 0, false, nil},
		{"Wrong", "***", false, 500, true, false, true, "",
			false, false, value.Value{}, false, false,
			10, 0, false, nil},
		{"Empty", "", false, 404, true, false, true, "",
			false, false, value.Value{}, false, false,
			10, 0, false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			var res response.Response
			var ch = make(chan response.Response)
			var wg sync.WaitGroup

			isSkipNotextMessages = tt.isSkipEmpty

			cc := NewChannels()
			cc.Add(tt.value)

			c, ok := cc.Find(tt.value)
			if !ok {
				c = &channel.Channel{Username: tt.value}
			}

			go func() {
				wg.Add(1)
				defer func() {
					if isPanic := recover() != nil; isPanic != tt.isPanic {
						t.Errorf("Получена паника: %v, ожидается: %v", isPanic, tt.isPanic)
					}
					wg.Done()
				}()
				Messages(ch, *c, tt.messagesCount, tt.afterMsgId, tt.isParseInfo)
			}()

			res = <-ch
			if res.Ok {
				*c = res.Data.(channel.Channel)
			}

			wg.Wait()

			if res.Ok != tt.result {
				t.Fatalf("Получен результат: %v, ожидается: %v", res.Ok, tt.result)
			}
			if res.Code != tt.code {
				t.Fatalf("Code - получено значение: %v, ожидается: %v", res.Code, tt.code)
			}
			if (res.Error != nil) != tt.isError && !tt.isPanic {
				t.Fatalf("Получена ошибка: %v, ожидается: %v", res.Error, tt.isError)
			}

			if !strings.Contains(c.Title, tt.title) {
				t.Errorf("Title - получено значение: %v, ожидается совпадение: %v", c.Title, tt.title)
			}
			if hasAbout := c.About != ""; hasAbout != tt.hasAbout {
				t.Errorf("About - получено значение: %v, ожидается: %v", hasAbout, tt.hasAbout)
			}

			if hasImage := c.Image != ""; hasImage != tt.hasImage {
				t.Errorf("Image - получено значение: %v, ожидается: %v", hasImage, tt.hasImage)
			}

			if !reflect.DeepEqual(c.Participants, tt.participants) {
				t.Errorf("Participants - получено значение: %#v, ожидается: %#v", c.Participants, tt.participants)
			}

			if c.IsVerified != tt.isVerified {
				t.Errorf("IsVerified - получено значение: %v, ожидается: %v", c.IsVerified, tt.isVerified)
			}
			if c.IsScam != tt.isScam {
				t.Errorf("IsScam - получено значение: %v, ожидается: %v", c.IsScam, tt.isScam)
			}

			for id, e := range tt.messages {
				var isFound bool
				for _, g := range c.Messages {
					if id != g.ID {
						continue
					}

					isFound = true

					if !strings.Contains(g.MessageHtml, e.MessageHtml) {
						t.Errorf("MessageHtml - получено значение: %q, ожидается совпадение: %q",
							g.MessageHtml, e.MessageHtml)
					}
					if !strings.Contains(g.MessageText, e.MessageText) {
						t.Errorf("MessageText - получено значение: %q, ожидается совпадение: %q",
							g.MessageText, e.MessageText)
					}

					if g.IsForwarded != e.IsForwarded {
						t.Errorf("IsForwarded - получено значение: %v, ожидается: %v",
							g.IsForwarded, e.IsForwarded)
					}
					if g.FwdLink != e.FwdLink {
						t.Errorf("FwdLink - получено значение: %v, ожидается: %v", g.FwdLink, e.FwdLink)
					}
					if g.FwdPost != e.FwdPost {
						t.Errorf("FwdPost - получено значение: %v, ожидается: %v", g.FwdPost, e.FwdPost)
					}
					if g.FwdTitle != e.FwdTitle {
						t.Errorf("FwdTitle - получено значение: %v, ожидается: %v", g.FwdTitle, e.FwdTitle)
					}
					if g.FwdAuthor != e.FwdAuthor {
						t.Errorf("FwdAuthor - получено значение: %v, ожидается: %v", g.FwdAuthor, e.FwdAuthor)
					}

					if g.IsEdited != e.IsEdited {
						t.Errorf("IsEdited - получено значение: %v, ожидается: %v", g.IsEdited, e.IsEdited)
					}
					if g.IsPoll != e.IsPoll {
						t.Errorf("IsPoll - получено значение: %v, ожидается: %v", g.IsPoll, e.IsPoll)
					}

					if g.HasImage != e.HasImage {
						t.Errorf("HasImage - получено значение: %v, ожидается: %v", g.HasImage, e.HasImage)
					}
					if g.HasVideo != e.HasVideo {
						t.Errorf("HasVideo - получено значение: %v, ожидается: %v", g.HasVideo, e.HasVideo)
					}
					if g.HasDocument != e.HasDocument {
						t.Errorf("HasDocument - получено значение: %v, ожидается: %v", g.HasDocument, e.HasDocument)
					}

					if e.Attachments != nil {
						var attMatches int
						for _, eAtt := range e.Attachments {
							for _, gAtt := range g.Attachments {
								if strings.Contains(gAtt, eAtt) {
									attMatches++
								}
							}
						}
						if attMatches != len(e.Attachments) {
							t.Errorf("Attachments - получены значения: %q, ожидаются совпадения: %q",
								g.Attachments, e.Attachments)
						}
					}
					if e.Hashtags != nil {
						var hashMatches int
						for _, eHash := range e.Hashtags {
							for _, gHash := range g.Hashtags {
								if strings.Contains(gHash, eHash) {
									hashMatches++
								}
							}
						}
						if hashMatches != len(e.Hashtags) {
							t.Errorf("Hashtags - получены значения: %q, ожидаются совпадения: %q",
								g.Hashtags, e.Hashtags)
						}
					}

					for _, link := range e.Buttons {
						if !g.Buttons.IsExist(link.Link) {
							t.Errorf("Buttons - получены значения: %#v, ожидаются: %#v, не найдено: %q",
								g.Buttons, e.Buttons, link.Link)
						}
					}
					for _, link := range e.Links {
						if !g.Links.IsExist(link.Link) {
							t.Errorf("Links - получены значения: %#v, ожидаются: %#v, не найдено: %q",
								g.Links, e.Links, link.Link)
						}
					}
					for _, link := range e.Advs {
						if !g.Advs.IsExist(link.Link) {
							t.Errorf("Advs - получены значения: %#v, ожидаются: %#v, не найдено: %q",
								g.Advs, e.Advs, link.Link)
						}
					}

					if e.Views.Value() != 0 {
						if !reflect.DeepEqual(g.Views, e.Views) {
							t.Errorf("Views - получено значение: %#v, ожидается: %#v", g.Views, e.Views)
						}
					}

					if !e.Date.IsZero() {
						if !g.Date.Equal(e.Date) {
							t.Errorf("Date - получено значение: %v, ожидается: %v", g.Date, e.Date)
						}
					}
					if !e.DateLocal.IsZero() {
						if !g.DateLocal.Equal(e.DateLocal) {
							t.Errorf("DateLocal - получено значение: %v, ожидается: %v", g.DateLocal, e.DateLocal)
						}
					}
				}

				if !isFound {
					t.Errorf("Сообщение не найдено (ID: %v)", id)
				}
			}
		})
	}
}

func TestPrintReport(t *testing.T) {
	valid := NewChannels()

	valid.Add("thecodemedia")
	valid.Add("codecamp")
	valid.Add("not_existed_channel")

	valid.Channels.Channels[0].Title = "Журнал «Код»"
	valid.Channels.Channels[1].Title = "CodeCamp"

	valid.Channels.Channels[0].Messages = []*message.Message{{ID: 1, MessageHtml: "A"}}

	empty := NewChannels()

	tests := []struct {
		test   string
		value  *Channels
		errs   []error
		result []string
	}{
		{"Valid", &valid, []error{errors.New("Канал @not_existed_channel не найден")},
			[]string{"2 (1)", "1 (2)", "Канал", "не найден", "Regexps", "---", "username"}},
		{"Empty", &empty, nil, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := mock.ReadOutput(func() {
				tt.value.PrintReport(tt.errs, true)
			})

			for _, match := range tt.result {
				if !strings.Contains(result, match) {
					t.Fatalf("Получено значение: %q, ожидаются совпадения: %q, не найдено: %q",
						result, tt.result, match)
				}
			}
		})
	}
}

func TestCountParsed(t *testing.T) {
	valid := NewChannels()

	valid.Add("thecodemedia")
	valid.Add("codecamp")
	valid.Add("not_existed_channel")

	valid.Channels.Channels[0].Title = "Журнал «Код»"
	valid.Channels.Channels[1].Title = "CodeCamp"

	empty := NewChannels()

	tests := []struct {
		test   string
		value  *Channels
		result int
	}{
		{"Valid", &valid, 2},
		{"Empty", &empty, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := tt.value.CountParsed()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestCountMessaged(t *testing.T) {
	valid := NewChannels()

	valid.Add("thecodemedia")
	valid.Add("codecamp")
	valid.Add("not_existed_channel")

	valid.Channels.Channels[0].Messages = []*message.Message{{ID: 1, MessageHtml: "A"}}
	valid.Channels.Channels[1].Messages = []*message.Message{{ID: 2, MessageHtml: "B"}}

	empty := NewChannels()

	tests := []struct {
		test   string
		value  *Channels
		result int
	}{
		{"Valid", &valid, 2},
		{"Empty", &empty, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := tt.value.CountMessaged()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestRemoveUnparsed(t *testing.T) {
	valid := NewChannels()

	valid.Add("thecodemedia")
	valid.Add("codecamp")
	valid.Add("not_existed_channel")

	valid.Channels.Channels[1].Title = "CodeCamp"

	empty := NewChannels()

	tests := []struct {
		test   string
		value  *Channels
		result int
		count  int
	}{
		{"Valid", &valid, 2, 1},
		{"Empty", &empty, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := tt.value.RemoveUnparsed()
			count := tt.value.Count()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
			if count != tt.count {
				t.Errorf("Получено количество: %v, ожидается: %v", count, tt.count)
			}
		})
	}
}

func TestRemoveUnmessaged(t *testing.T) {
	valid := NewChannels()

	valid.Add("thecodemedia")
	valid.Add("codecamp")

	valid.Channels.Channels[0].Messages = []*message.Message{{ID: 1, MessageHtml: "A"}}

	empty := NewChannels()

	tests := []struct {
		test   string
		value  *Channels
		result int
		count  int
	}{
		{"Valid", &valid, 1, 1},
		{"Empty", &empty, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := tt.value.RemoveUnmessaged()
			count := tt.value.Count()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
			if count != tt.count {
				t.Errorf("Получено количество: %v, ожидается: %v", count, tt.count)
			}
		})
	}
}

func TestTestProxy(t *testing.T) {
	empty := NewChannels()

	tests := []struct {
		test   string
		value  *Channels
		step   uint
		limit  uint
		pause  time.Duration
		result []string
	}{
		{"Empty", &empty, 1, 5, 0, []string{"1:", "2:", "3:", "4:", "5:", "-", "0%"}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := mock.ReadOutput(func() {
				tt.value.TestProxy(tt.step, tt.limit, tt.pause)
			})

			for _, match := range tt.result {
				if !strings.Contains(result, match) {
					t.Fatalf("Получено значение: %q, ожидаются совпадения: %q, не найдено: %q",
						result, tt.result, match)
				}
			}
		})
	}
}

func TestSplitMessages(t *testing.T) {
	text := "...<div class=\"tgme_widget_message_wrap\">\nMessage 1\n</div>" +
		"<div class=\"tgme_widget_message_wrap message\">\nMessage 2\n</div>" +
		"<div class=\"tgme_widget message_wrap messagewrap\">\nMessage 3\n</div>" +
		"<div class=\"tgme_widget_message_wrap message-wrap\">\nMessage 4\n</div>..."

	tests := []struct {
		test   string
		value  string
		result int
	}{
		{"Many", text, 3},
		{"One", "Message", 1},
		{"Empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := len(splitMessages(tt.value))

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestSplitMessagesRegexp(t *testing.T) {
	text := "...<div class=\"tgme_widget_message_wrap\">\nMessage 1\n</div>" +
		"<div class=\"tgme_widget_message_wrap message\">\nMessage 2\n</div>" +
		"<div class=\"tgme_widget message_wrap messagewrap\">\nMessage 3\n</div>" +
		"<div class=\"tgme_widget_message_wrap message-wrap\">\nMessage 4\n</div>..."

	tests := []struct {
		test   string
		value  string
		result int
	}{
		{"Many", text, 2},
		{"One", "Message", 0},
		{"Empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := len(splitMessagesRegexp(tt.value))

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}
