package channel

import (
	"strings"
	"testing"
	"time"

	"statosphere/parser/links"
	"statosphere/parser/message"
	"statosphere/parser/mock"
	"statosphere/parser/value"
)

func TestPrint(t *testing.T) {
	channelUsername := Channel{
		Username:     "username",
		Peer:         "@username",
		Link:         "https://t.me/username",
		Title:        "Channel",
		About:        "About info",
		Contacts:     links.New(),
		Siblings:     links.New(),
		Image:        "image.jpg",
		Kind:         "channel",
		Participants: value.Value{Approx: 100000, Short: "0.1M"},
		Photos:       value.Value{Approx: 1500, Short: "1.5K"},
		Videos:       value.Value{Exact: 5},
		Links:        value.Value{Exact: 500},
		IsVerified:   true,
	}

	channelUsername.Contacts.Add("@username", "Username", 0)
	channelUsername.Contacts.Add("@sibling", "", 0)
	channelUsername.Contacts.Add("username@mail.com", "Mail", 0)
	channelUsername.Siblings.Add("https://t.me/sibling", "", 0)

	resultUsername := []string{
		"Username: username",
		"Peer: @username",
		"Link: https://t.me/username",
		"Title: Channel",
		`About: "About info"`,
		"Contacts:",
		"@username [Username]",
		"@sibling",
		"username@mail.com [Mail]",
		"Siblings:",
		"https://t.me/sibling",
		"Image: +",
		"Kind: channel",
		"Participants: 100000 (0.1M)",
		"Photos: 1500 (1.5K)",
		"Videos: 5",
		"Links: 500",
		"Is verified: true",
		"Is scam: false",
		"Messages: 0",
	}

	channelJoinchat := Channel{
		Joinchat:     "abc4_fGhI0-LmnOp",
		Peer:         "+abc4_fGhI0-LmnOp",
		Link:         "https://t.me/joinchat/abc4_fGhI0-LmnOp",
		Title:        "Private channel",
		About:        "Secret about info",
		Kind:         "private",
		Participants: value.Value{Exact: 555},
		IsScam:       true,
	}

	resultJoinchat := []string{
		"Joinchat: abc4_fGhI0-LmnOp",
		"Peer: +abc4_fGhI0-LmnOp",
		"Link: https://t.me/joinchat/abc4_fGhI0-LmnOp",
		"Title: Private channel",
		`About: "Secret about info"`,
		"Kind: private",
		"Participants: 555",
		"Is verified: false",
		"Is scam: true",
	}

	channelUser := Channel{
		Username: "tguser",
		Peer:     "@tguser",
		Link:     "https://t.me/tguser",
		Title:    "My name",
		About:    "My info",
		Kind:     "user",
	}

	resultUser := []string{
		"Username: tguser",
		"Peer: @tguser",
		"Link: https://t.me/tguser",
		"Title: My name",
		`About: "My info"`,
		"Kind: user",
	}

	tests := []struct {
		test    string
		channel Channel
		result  []string
	}{
		{"Username", channelUsername, resultUsername},
		{"Joinchat", channelJoinchat, resultJoinchat},
		{"User", channelUser, resultUser},
		{"Empty", Channel{}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := mock.ReadOutput(func() {
				tt.channel.Print(true)
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

func TestPrintMessages(t *testing.T) {
	channel := Channel{}

	channel.Messages = append(channel.Messages, &message.Message{
		ID:          5,
		MessageHtml: "Simple message",
		Views:       value.Value{Approx: 3500, Short: "3.5K"},
		Date:        time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC),
		DateLocal:   time.Date(2006, time.January, 2, 18, 4, 5, 0, time.Local),
	})

	channel.Messages = append(channel.Messages, &message.Message{
		ID:          10,
		MessageHtml: "Message with #hashtag",
		IsEdited:    true,
		Hashtags:    []string{"hashtag"},
		Views:       value.Value{Exact: 250},
		Date:        time.Date(2007, time.January, 2, 15, 4, 5, 0, time.UTC),
		DateLocal:   time.Date(2007, time.January, 2, 18, 4, 5, 0, time.Local),
	})

	result := []string{
		"ID: 5",
		"Views: 3500 (3.5K)",
		"Date: 2006-01-02 15:04:05",
		"Text:",
		"Simple message",
		"---",
		"ID: 10",
		"Views: 250",
		"Date: 2007-01-02 15:04:05",
		"Hashtags: [hashtag]",
		"Text (edited):",
		"Message with #hashtag",
	}

	tests := []struct {
		test    string
		channel Channel
		result  []string
	}{
		{"Messages", channel, result},
		{"Empty", Channel{}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := mock.ReadOutput(func() {
				tt.channel.PrintMessages()
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
