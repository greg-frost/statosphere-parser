package message

import (
	"strings"
	"testing"
	"time"

	"statosphere/parser/links"
	"statosphere/parser/mock"
	"statosphere/parser/value"
)

func TestPrint(t *testing.T) {
	messageLinks := Message{
		ID:          1,
		MessageHtml: `Message with <a href="@username">Username</a>`,
		MessageText: "Message with Username",
		HasImage:    true,
		Attachments: []string{"img.jpg", "img.png"},
		Links:       links.New(),
		Advs:        links.New(),
		Views:       value.Value{Approx: 1000, Short: "1K"},
		Date:        time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC),
		DateLocal:   time.Date(2006, time.January, 2, 18, 4, 5, 0, time.Local),
	}

	messageLinks.Links.Add("@username", "Username", 0)
	messageLinks.Advs.Add("https://t.me/username", "", 0)

	resultLinks := []string{
		"ID: 1",
		"Views: 1000 (1K)",
		"Date: 2006-01-02 15:04:05",
		"Image: img.jpg (2)",
		"Links:",
		"@username [Username]",
		"Advs:",
		"https://t.me/username",
		"Text:",
		`Message with <a href="@username">Username</a>`,
	}

	messageForward := Message{
		ID:          2,
		MessageHtml: `Forwarded message with <b class="bold">tag</b>`,
		MessageText: "Forwarded message with tag",
		IsForwarded: true,
		FwdLink:     "https://t.me/username",
		FwdPost:     100,
		FwdTitle:    "My channel",
		FwdAuthor:   "Username",
		HasVideo:    true,
		Attachments: []string{"video.flv"},
		Views:       value.Value{Approx: 25000, Short: "25K"},
		Date:        time.Date(2007, time.January, 2, 15, 4, 5, 0, time.UTC),
		DateLocal:   time.Date(2007, time.January, 2, 18, 4, 5, 0, time.Local),
	}

	resultForward := []string{
		"ID: 2",
		"Forwarded: My channel (Username)",
		"https://t.me/username/100",
		"Views: 25000 (25K)",
		"Date: 2007-01-02 15:04:05",
		"Video: video.flv",
		"Text:",
		`Forwarded message with <b class="bold">tag</b>`,
	}

	messageHashtags := Message{
		ID:          3,
		MessageHtml: "Message with #link and #tag",
		MessageText: "Message with #link and #tag",
		IsEdited:    true,
		HasDocument: true,
		Attachments: []string{"file.pdf", "file.doc", "file.xls"},
		Hashtags:    []string{"link", "tag"},
		Views:       value.Value{Exact: 755},
		Date:        time.Date(2008, time.January, 2, 15, 4, 5, 0, time.UTC),
		DateLocal:   time.Date(2008, time.January, 2, 18, 4, 5, 0, time.Local),
	}

	resultHashtags := []string{
		"ID: 3",
		"Views: 755",
		"Date: 2008-01-02 15:04:05",
		"Document: file.pdf (3)",
		"Hashtags: [link tag]",
		"Text (edited):",
		"Message with #link and #tag",
	}

	tests := []struct {
		test    string
		message Message
		result  []string
	}{
		{"Links", messageLinks, resultLinks},
		{"Forward", messageForward, resultForward},
		{"Hashtags", messageHashtags, resultHashtags},
		{"Empty", Message{}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := mock.ReadOutput(func() {
				tt.message.Print()
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
