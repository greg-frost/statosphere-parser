package message

import (
	"fmt"
	"time"

	"statosphere/parser/format"
	"statosphere/parser/links"
	"statosphere/parser/value"
)

// Сообщение
type Message struct {
	ID          uint        `json:"id"`
	MessageHtml string      `json:"message,omitempty"`
	MessageText string      `json:"-"`
	IsForwarded bool        `json:"isForwarded"`
	FwdLink     string      `json:"forwardedLink,omitempty"`
	FwdPost     uint        `json:"forwardedPost,omitempty"`
	FwdTitle    string      `json:"forwardedTitle,omitempty"`
	FwdAuthor   string      `json:"forwardedAuthor,omitempty"`
	IsEdited    bool        `json:"isEdited"`
	IsPoll      bool        `json:"isPoll"`
	HasImage    bool        `json:"hasImage"`
	HasVideo    bool        `json:"hasVideo"`
	HasDocument bool        `json:"hasDocument"`
	Attachments []string    `json:"attachments,omitempty"`
	Hashtags    []string    `json:"hashtags,omitempty"`
	Buttons     links.Links `json:"buttons,omitempty"`
	Links       links.Links `json:"links,omitempty"`
	Advs        links.Links `json:"advs,omitempty"`
	Views       value.Value `json:"views"`
	Date        time.Time   `json:"date"`
	DateLocal   time.Time   `json:"-"`
}

// Печать сообщения
func (m *Message) Print() {
	if m.ID == 0 {
		return
	}

	fmt.Printf("\n-----\n\n")

	// ID
	fmt.Println("ID:", m.ID)

	// Репост
	if m.IsForwarded {
		fmt.Print("Forwarded: ", format.Truncate(m.FwdTitle, 20))
		if m.FwdAuthor != "" {
			fmt.Printf(" (%s)", format.Truncate(m.FwdAuthor, 10))
		}
		fmt.Println()

		if m.FwdLink != "" {
			fmt.Print("   " + m.FwdLink)
			if m.FwdPost > 0 {
				fmt.Print("/", m.FwdPost)
			}
			fmt.Println()
		}
	}

	// Число просмотров
	fmt.Println("Views:", m.Views)

	// Дата публикации
	fmt.Println("Date:", format.Date(m.Date))

	// Медиа
	if len(m.Attachments) > 0 {
		switch {
		case m.HasImage:
			fmt.Print("Image: ")
		case m.HasVideo:
			fmt.Print("Video: ")
		case m.HasDocument:
			fmt.Print("Document: ")
		}
		fmt.Print(format.Truncate(m.Attachments[0], 30))
		if len(m.Attachments) > 1 {
			fmt.Printf(" (%d)", len(m.Attachments))
		}
		fmt.Println()
	}

	// Хэштеги
	if len(m.Hashtags) > 0 {
		fmt.Println("Hashtags:", m.Hashtags)
	}

	// Ссылки
	m.Links.Print("\nLinks:")

	// Рекламные ссылки
	m.Advs.Print("\nAdvs:")

	// Текст
	if m.MessageHtml != "" {
		var editedMark string
		if m.IsEdited {
			editedMark = " (edited)"
		}

		fmt.Printf("\nText%s:\n\n", editedMark)
		fmt.Println(m.MessageHtml)
	}
}
