package channel

import (
	"fmt"

	"statosphere/parser/links"
	"statosphere/parser/message"
	"statosphere/parser/value"
)

// Канал
type Channel struct {
	Username     string             `json:"username,omitempty"`
	Joinchat     string             `json:"joinchat,omitempty"`
	Peer         string             `json:"peer"`
	Link         string             `json:"link"`
	Title        string             `json:"title"`
	About        string             `json:"about,omitempty"`
	Contacts     links.Links        `json:"contacts,omitempty"`
	Siblings     links.Links        `json:"siblings,omitempty"`
	Image        string             `json:"image,omitempty"`
	Kind         string             `json:"kind"`
	Participants value.Value        `json:"participants"`
	Photos       value.Value        `json:"photos,omitempty"`
	Videos       value.Value        `json:"videos,omitempty"`
	Files        value.Value        `json:"files,omitempty"`
	Links        value.Value        `json:"links,omitempty"`
	IsVerified   bool               `json:"isVerified"`
	IsScam       bool               `json:"isScam"`
	Messages     []*message.Message `json:"messages,omitempty"`
}

// Печать канала
func (c *Channel) Print(isPrintMessages bool) {
	if c.Title == "" {
		return
	}

	// Юзернейм, джойнчат, краткая и полная ссылки
	fmt.Println("Username:", c.Username)
	fmt.Println("Joinchat:", c.Joinchat)
	fmt.Println("Peer:", c.Peer)
	fmt.Println("Link:", c.Link)

	// Название
	fmt.Println("Title:", c.Title)

	// Описание
	fmt.Printf("About: %q\n", c.About)

	// Контактные ссылки
	c.Contacts.Print("Contacts:")

	// Родственные ссылки
	c.Siblings.Print("Siblings:")

	// Изображение
	fmt.Print("Image:")
	if c.Image != "" {
		fmt.Print(" + ")
	}
	fmt.Println()

	// Тип
	fmt.Println("Kind:", c.Kind)

	// Число подписчиков
	if c.Kind == "channel" || c.Kind == "private" || c.Kind == "chat" {
		fmt.Println("Participants:", c.Participants)
	}

	// Количество фото, видео, файлов и ссылок
	if c.Kind == "channel" {
		fmt.Println("Photos:", c.Photos)
		fmt.Println("Videos:", c.Videos)
		fmt.Println("Files:", c.Files)
		fmt.Println("Links:", c.Links)
	}

	// Метки верификации и скама
	fmt.Println("Is verified:", c.IsVerified)
	fmt.Println("Is scam:", c.IsScam)

	// Сообщения
	if c.Kind == "channel" {
		fmt.Println("Messages:", len(c.Messages))

		if isPrintMessages {
			c.PrintMessages()
		}
	}

	fmt.Print("\n***\n\n")
}

// Печать сообщений канала
func (c *Channel) PrintMessages() {
	if len(c.Messages) == 0 {
		return
	}

	for _, m := range c.Messages {
		m.Print()
	}
}
