package channels

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"statosphere/parser/channel"
	"statosphere/parser/message"
	"statosphere/parser/mock"
)

var (
	validUsernames       = []string{"username", "@username", "t.me/username"}
	partlyValidUsernames = []string{"@user", "username.t.me", "https://t.me/username"}
	invalidUsernames     = []string{"", "@user", "t.me/joinchat/username"}
)

func TestNew(t *testing.T) {
	tests := []struct {
		test   string
		result Channels
	}{
		{"Valid", Channels{Channels: make([]*channel.Channel, 0)}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := New()

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		test   string
		value  string
		result bool
	}{
		{"Username", "username", true},
		{"UsernameShort", "user", false},
		{"UsernameAt", "@username", true},
		{"UsernameTme", "t.me/username", true},
		{"UsernameHttp", "http://t.me/username", true},
		{"UsernameDomain", "https://username.t.me", true},
		{"UsernameResolve", "tg://resolve?domain=username", true},
		{"JoinchatPlus", "+abc4_fGhI0-LmnOp", true},
		{"JoinchatPrefix", "joinchat/abc4_fGhI0-LmnOp", true},
		{"JoinchatTme", "t.me/+abc4_fGhI0-LmnOp", true},
		{"JoinchatShort", "t.me/joinchat/abc4_fGhI0-Lm", false},
		{"JoinchatInvite", "tg://join?invite=abc4_fGhI0-LmnOp", true},
		{"JoinchatChars", "joinchat/AAAAAEab3D*EfGh0+KLmnO", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			channels := New()
			result := channels.Add(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPrepare(t *testing.T) {
	tests := []struct {
		test   string
		values []string
		result int
	}{
		{"Valid", validUsernames, 3},
		{"PartlyValid", partlyValidUsernames, 2},
		{"Invalid", invalidUsernames, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			channels := New()
			result := channels.PrepareFromList(tt.values)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPrepareFromString(t *testing.T) {
	tests := []struct {
		test   string
		values string
		result int
	}{
		{"Valid", strings.Join(validUsernames, ", "), 3},
		{"PartlyValid", "[" + strings.Join(partlyValidUsernames, ",") + "]", 2},
		{"Invalid", " { " + strings.Join(invalidUsernames, " , ") + "}", 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			channels := New()
			result := channels.PrepareFromString(tt.values)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPrepareFromList(t *testing.T) {
	tests := []struct {
		test   string
		values []string
		result int
	}{
		{"Valid", validUsernames, 3},
		{"PartlyValid", partlyValidUsernames, 2},
		{"Invalid", invalidUsernames, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			channels := New()
			result := channels.PrepareFromList(tt.values)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPrepareFromFile(t *testing.T) {
	tests := []struct {
		test   string
		values []string
		result int
	}{
		{"Valid", validUsernames, 3},
		{"PartlyValid", partlyValidUsernames, 2},
		{"Invalid", invalidUsernames, 0},
	}

	for i, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			filename := "channels_temp_" + fmt.Sprint(i)
			file, _ := os.Create(filename)
			for _, proxy := range tt.values {
				file.WriteString(proxy + "\n\n")
			}
			defer os.Remove(filename)
			defer file.Close()

			channels := New()
			result := channels.PrepareFromFile(filename)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestFind(t *testing.T) {
	channels := New()
	channels.Add("username1")
	channels.Add("@username2")
	channels.Add("t.me/username3")

	tests := []struct {
		test    string
		value   string
		result  *channel.Channel
		isExist bool
	}{
		{"FoundFirst", "@username1", channels.Channels[0], true},
		{"FoundSecond", "t.me/username2", channels.Channels[1], true},
		{"NotFound", "nonexisted", nil, false},
		{"Empty", "", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result, isExist := channels.Find(tt.value)

			if isExist != tt.isExist {
				t.Fatalf("Получен результат: %v, ожидается: %v", isExist, tt.isExist)
			}
			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestList(t *testing.T) {
	username := "https://t.me/username"

	tests := []struct {
		test   string
		values []string
		result []string
	}{
		{"Valid", validUsernames, []string{username, username, username}},
		{"PartlyValid", partlyValidUsernames, []string{username, username}},
		{"Invalid", invalidUsernames, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			channels := New()
			channels.Prepare(tt.values)

			result := channels.List()

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestLimit(t *testing.T) {
	tests := []struct {
		test   string
		from   uint
		to     uint
		result int
	}{
		{"None", 0, 5, 5},
		{"From", 3, 5, 2},
		{"To", 0, 4, 4},
		{"Range", 2, 4, 2},
		{"OutOfRange", 0, 10, 5},
		{"SwapRange", 4, 2, 0},
		{"Clear", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			channels := New()
			channels.Prepare(validUsernames)
			channels.Prepare(partlyValidUsernames)

			channels.Limit(tt.from, tt.to)
			result := channels.Count()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		test   string
		values []string
		result int
	}{
		{"Valid", validUsernames, 3},
		{"PartlyValid", partlyValidUsernames, 2},
		{"Invalid", invalidUsernames, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			channels := New()
			channels.Prepare(tt.values)

			result := channels.Count()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	channels := New()

	channels.Channels = append(channels.Channels, &channel.Channel{
		Username:   "username",
		Title:      "Channel",
		About:      "About info",
		Kind:       "channel",
		IsVerified: true,
	})

	channels.Channels = append(channels.Channels, &channel.Channel{
		Peer:  "@tguser",
		Title: "My name",
		Kind:  "user",
	})

	result := []string{
		"Channels",
		"---",
		"Username: username",
		"Title: Channel",
		`About: "About info"`,
		"Kind: channel",
		"Is verified: true",
		"***",
		"Peer: @tguser",
		"Title: My name",
		"Kind: user",
	}

	tests := []struct {
		test     string
		channels Channels
		result   []string
	}{
		{"Channels", channels, result},
		{"Empty", Channels{}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := mock.ReadOutput(func() {
				tt.channels.Print(true)
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
	channels := New()

	channels.Channels = append(channels.Channels, &channel.Channel{
		Link:  "https://t.me/username",
		Title: "Channel",
		Kind:  "channel",
	})

	channels.Channels[0].Messages = append(channels.Channels[0].Messages, &message.Message{
		ID:          5,
		MessageHtml: "Simple message",
	})

	channels.Channels[0].Messages = append(channels.Channels[0].Messages, &message.Message{
		ID:          10,
		MessageHtml: "Message with #hashtag",
		IsEdited:    true,
		Hashtags:    []string{"hashtag"},
	})

	channels.Channels = append(channels.Channels, &channel.Channel{
		Peer:  "@tguser",
		Title: "My name",
		Kind:  "user",
	})

	result := []string{
		"Messages",
		"[ https://t.me/username ]",
		"---",
		"ID: 5",
		"Text:",
		"Simple message",
		"---",
		"ID: 10",
		"Hashtags: [hashtag]",
		"Text (edited):",
		"Message with #hashtag",
	}

	tests := []struct {
		test     string
		channels Channels
		result   []string
	}{
		{"Messages", channels, result},
		{"Empty", Channels{}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := mock.ReadOutput(func() {
				tt.channels.PrintMessages()
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

func TestRemove(t *testing.T) {
	channels := New()
	channels.Prepare(validUsernames)

	tests := []struct {
		test   string
		value  string
		result bool
		count  int
	}{
		{"NotFound", "nonexisted", false, 3},
		{"Delete", "username", true, 2},
		{"Empty", "", false, 2},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := channels.Remove(tt.value)
			count := channels.Count()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
			if count != tt.count {
				t.Errorf("Получено количество: %v, ожидается: %v", count, tt.count)
			}
		})
	}
}

func TestRemoveByIdx(t *testing.T) {
	channels := New()
	channels.Prepare(validUsernames)

	tests := []struct {
		test   string
		value  int
		result bool
		count  int
	}{
		{"NotFound", 5, false, 3},
		{"Delete", 2, true, 2},
		{"DeletedAlready", 2, false, 2},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := channels.RemoveByIdx(tt.value)
			count := channels.Count()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
			if count != tt.count {
				t.Errorf("Получено количество: %v, ожидается: %v", count, tt.count)
			}
		})
	}
}

func TestRemoveByLink(t *testing.T) {
	channels := New()
	channels.Prepare(validUsernames)

	tests := []struct {
		test   string
		value  string
		result bool
		count  int
	}{
		{"NotFound", "nonexisted", false, 3},
		{"Delete", "username", true, 2},
		{"DeleteAgain", "username", true, 1},
		{"Empty", "", false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := channels.RemoveByLink(tt.value)
			count := channels.Count()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
			if count != tt.count {
				t.Errorf("Получено количество: %v, ожидается: %v", count, tt.count)
			}
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	channels := New()
	channels.Prepare(validUsernames)

	tests := []struct {
		test     string
		isUnique bool
		result   int
	}{
		{"None", false, 3},
		{"Unique", true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isUnique {
				channels.RemoveDuplicates()
			}

			result := channels.Count()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestClear(t *testing.T) {
	channels := New()
	channels.Prepare(validUsernames)

	tests := []struct {
		test    string
		isClear bool
		result  int
	}{
		{"None", false, 3},
		{"Clear", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isClear {
				channels.Clear()
			}

			result := channels.Count()

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}
