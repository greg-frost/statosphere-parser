package parse

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"statosphere/parser/cache"
	"statosphere/parser/channel"
	"statosphere/parser/channels"
	"statosphere/parser/check"
	"statosphere/parser/date"
	"statosphere/parser/format"
	"statosphere/parser/get"
	"statosphere/parser/links"
	"statosphere/parser/message"
	"statosphere/parser/proxy"
	"statosphere/parser/regexp"
	"statosphere/parser/response"
	"statosphere/parser/value"
)

// Набор каналов для парсинга
type Channels struct {
	channels.Channels
	m sync.Mutex
}

// Конструктор набора каналов для парсинга
func NewChannels() Channels {
	return Channels{
		Channels: channels.New(),
	}
}

// Парсинг каналов
func (pc *Channels) Parse(ctx context.Context, isExactParticipants bool, messagesCount uint) (int, []error) {
	const cacheDuration = 5 * time.Minute // Время кеширования

	var (
		parsed int
		errs   []error
		wg     sync.WaitGroup
		m      sync.Mutex
	)

	channelsCount := pc.Count()

	if channelsCount == 0 {
		errs = append(errs, fmt.Errorf("Ошибка парсинга: %w", errors.New("нет каналов")))
	}

	for i, c := range pc.Channels.Channels {

		// Попытка чтения из кэша
		if value, ok := cache.Value(c.Link); ok {
			cacheChannel := value.(*channel.Channel)
			cacheMessagesCount := uint(len(cacheChannel.Messages))

			if messagesCount == cacheMessagesCount {
				pc.Channels.Channels[i] = cacheChannel
				parsed++
				continue
			} else {
				cache.Remove(c.Link)
			}
		}

		// Параллельный парсинг каналов

		wg.Add(1)

		go func(c *channel.Channel) {
			defer wg.Done()

			var resInfo, resMsgs response.Response
			var chInfo, chMsgs = make(chan response.Response), make(chan response.Response)

			// Если нужно точное число подпичсиков
			if isExactParticipants {
				go Info(chInfo, *c)
			} else {
				go func() { chInfo <- response.Response{} }()
			}

			// Если нужны сообщения или достаточно приближенного числа подписчиков
			if c.Username != "" && (messagesCount > 0 || !isExactParticipants) {
				go Messages(chMsgs, *c, messagesCount, 0, true)
			} else {
				go func() { chMsgs <- response.Response{} }()
			}

			// Ожидание всех результатов
			select {
			case resInfo = <-chInfo:
				resMsgs = <-chMsgs
			case <-ctx.Done():
				errs = append(errs, fmt.Errorf("Ошибка парсинга: %w", fmt.Errorf("отмена для %v", c.Peer)))
				return
			}

			pc.m.Lock()
			defer pc.m.Unlock()

			// Обработка полученных данных

			if resInfo.Ok {
				infoData := resInfo.Data.(channel.Channel)
				*c = infoData
			}

			if resMsgs.Ok {
				msgsData := resMsgs.Data.(channel.Channel)

				if !resInfo.Ok {
					*c = msgsData
				} else {
					c.Photos = msgsData.Photos
					c.Videos = msgsData.Videos
					c.Files = msgsData.Files
					c.Links = msgsData.Links
				}

				if len(msgsData.Messages) > 0 {
					c.Messages = msgsData.Messages
				}
			}

			// Обработка результатов и ошибок

			m.Lock()
			defer m.Unlock()

			if resInfo.Ok || resMsgs.Ok {
				cache.SetValue(c.Link, c, cacheDuration) // Запись в кэш
				parsed++
			}

			if resInfo.Error != nil {
				errs = append(errs, fmt.Errorf("Ошибка парсинга информации: %w", resInfo.Error))
			}

			if resMsgs.Error != nil {
				errs = append(errs, fmt.Errorf("Ошибка парсинга сообщений: %w", resMsgs.Error))
			}
		}(c)
	}

	wg.Wait()

	return parsed, errs
}

// Печать отчета
func (pc *Channels) PrintReport(errs []error, isPrintRegexpReport bool) {
	if errs != nil {
		for _, e := range errs {
			fmt.Println(e)
		}
		fmt.Println()
	}

	count := pc.Count()
	parsed := pc.CountParsed()
	messaged := pc.CountMessaged()

	fmt.Printf("Успешно: %d (%d)\nНеудачно: %d (%d)\n\n", parsed, messaged, count-parsed, count-messaged)

	if isPrintRegexpReport {
		regexp.ProfilerReport(0 * time.Millisecond)
	}
}

// Количество спарсенных каналов
func (pc *Channels) CountParsed() (count int) {
	for _, c := range pc.Channels.Channels {
		if c.Title != "" {
			count++
		}
	}
	return count
}

// Количество каналов с сообщениями
func (pc *Channels) CountMessaged() (count int) {
	for _, c := range pc.Channels.Channels {
		if len(c.Messages) > 0 {
			count++
		}
	}
	return count
}

// Удаление неспарсенных каналов
func (pc *Channels) RemoveUnparsed() (removed int) {
	count := pc.Count()

	for i := 0; i < count; i++ {
		if pc.Channels.Channels[i].Title == "" {
			pc.RemoveByIdx(i)
			removed++
			count--
			i--
		}
	}

	return removed
}

// Удаление каналов без сообщений
func (pc *Channels) RemoveUnmessaged() (removed int) {
	count := pc.Count()

	for i := 0; i < count; i++ {
		if len(pc.Channels.Channels[i].Messages) == 0 {
			pc.RemoveByIdx(i)
			removed++
			count--
			i--
		}
	}

	return removed
}

// Тестирование прокси
func (pc *Channels) TestProxy(step, limit uint, pause time.Duration) {
	var (
		i     uint
		count int
		total int
		m     sync.Mutex
		wg    sync.WaitGroup
	)

	pc.Limit(0, step)

	for i = 0; i < limit; i = i + step {
		wg.Add(1)
		count++

		go func(count int) {
			defer wg.Done()

			proxy := proxy.Current()
			if proxy == "" {
				proxy = "-"
			}

			parsed, _ := pc.Parse(context.Background(), true, 0)

			fmt.Printf("%2d: %3d  %s\n", count, parsed, proxy)

			m.Lock()
			defer m.Unlock()

			total += parsed
		}(count)

		time.Sleep(pause)
	}

	wg.Wait()

	fmt.Printf("\nУспешно: %d / %d (%0.0f%%)\n", total, limit, float64(total)/float64(limit)*100)
}

// Общие паттерны для парсинга
const (
	PatternIsValidInfo  = `tgme_page_(title|extra)`
	PatternIsValidMsgs  = `tgme_header_(title|counter)`
	PatternIsntEmpty    = `meta.*?robots.*?no(index|follow)`
	PatternTitleProp    = `property="(og|twitter):title"[^>]+?content="(?P<title>[^"]+?)"`
	PatternTitleBody    = `tgme_(page|(channel_info_)?header)_title[^>]+?>([^>]+?>)?(?P<title>.+?)</div`
	PatternAboutProp    = `property="(og|twitter):description"[^>]+?content="(?P<about>[^"]*?)"`
	PatternAboutBody    = `tgme_(page|channel_info)_description[^>]+?>(?P<about>.*?)</div`
	PatternImageProp    = `property="(og|twitter):image".+?content="(?P<image>.+?` + PatternImage + `)?"`
	PatternImageBody    = `tgme_page_photo_image.+?src="(?P<image>.+?` + PatternImage + `)?"`
	PatternScamEn       = `warning.+?report.+?(scam|fake).+?account.+?careful.+?money`
	PatternScamRu       = `внимание.+?жалова.+?(мошенничество|выдать себя).+?аккаунт.+?осторожн.+?ден[еь]г`
	PatternVerified     = `<i class="verified-icon">`
	PatternNumber       = `[\d\s\.,_]+`
	PatternFactorNumber = `[\d\s\.,_MK]+`
	PatternImage        = `\.(jpe?g|png)`
	PatternVideo        = `\.(mp(eg-?)?4|flv)`
)

// Парсинг основной информации канала
func Info(chRes chan<- response.Response, c channel.Channel) {
	defer func() {
		if err := recover(); err != nil {
			chRes <- response.Response{Ok: false, Code: 500, Error: err.(error)}
			log.Panicln("Паника парсинга информации:", err)
		}
	}()

	channelPeer, channelLink := format.Username(c.Username, c.Joinchat, 0)
	infoLink, _ := format.PageLink(c.Username, c.Joinchat)

	// Получение страницы
	code, page, err := get.Page(infoLink)
	if code != 200 || err != nil {
		chRes <- response.Response{Ok: false, Code: code, Error: err}
		return
	}

	// Если страница невалидна
	re := regexp.Prepare("isValid", PatternIsValidInfo)
	if !re.Match(page) {
		chRes <- response.Response{Ok: false, Code: 404, Error: fmt.Errorf("нет данных для %s", channelPeer)}
		return
	}

	// ... или пуста
	re = regexp.Prepare("isntEmpty", PatternIsntEmpty)
	if re.Match(page) {
		chRes <- response.Response{Ok: false, Code: 404, Error: fmt.Errorf("нет данных для %s", channelPeer)}
		return
	}

	c.Contacts = links.New()
	c.Siblings = links.New()

	// Название

	patterns := []string{
		PatternTitleProp,
		`(?s)` + PatternTitleBody,
		`(?s)tgme_page_additional.+?(join|contact).+?>(?P<title>.+?)</`,
	}

	for _, pattern := range patterns {
		re = regexp.Prepare("title", pattern)
		res := re.Find(page)

		if len(res) > 0 {
			c.Title = format.SafeString(res["title"])
			break
		}
	}

	if c.Title == "" {
		panic(fmt.Errorf("не получено название для %s", channelPeer))
	}

	// Описание

	patterns = []string{
		`(?s)` + PatternAboutBody,
		`(?s)` + PatternAboutProp,
	}

	for _, pattern := range patterns {
		re = regexp.Prepare("about", pattern)
		res := re.Find(page)

		if len(res) > 0 {
			c.About = format.SafeText(res["about"])
			break
		}
	}

	re = regexp.Prepare("isntAbout", `(view|join|contact).+?right`)
	if re.Match(c.About) {
		c.About = ""
	}

	// Изображение

	patterns = []string{
		PatternImageBody,
		PatternImageProp,
	}

	for _, pattern := range patterns {
		re = regexp.Prepare("image", pattern)
		res := re.Find(page)

		if len(res) > 0 {
			c.Image = res["image"]
			break
		}
	}

	// Число подписчиков

	pattern := `tgme_page_extra.+?>(?P<participants>` + PatternNumber + `).*?</`

	re = regexp.Prepare("participants", pattern)
	res := re.Find(page)

	c.Participants, err = value.New(res["participants"], true)

	if res["participants"] != "" && err != nil {
		panic(fmt.Errorf("не получено число подписчиков для %s", channelPeer))
	}

	// Кнопка

	pattern = `tgme_action_button.+?>(?P<button>.+?)</`

	re = regexp.Prepare("button", pattern)
	res = re.Find(page)

	button := res["button"]

	// Тип

	patterns = []string{
		`tgme_page_context_link.+?href="/s/.+?"`,
		`(?i)join.+?(channel|(super)?group)`,
		`(?i)bot(father)?$`,
		`(?i)send.+?message`,
	}

	kinds := []string{
		"channel",
		"private",
		"bot",
		"user",
	}

	for i, pattern := range patterns {
		kind := kinds[i]

		var where string
		switch kind {
		case "channel":
			where = page
		case "private", "user":
			where = button
		case "bot":
			where = c.Username
		}

		if !((kind == "user" || kind == "bot") && c.Participants.Value() != 0) {
			re = regexp.Prepare(kind, pattern)
			if re.Match(where) {
				c.Kind = kind
				break
			}
		}
	}

	if c.Kind == "" {
		c.Kind = "chat"
	}

	if (c.Kind == "channel" || c.Kind == "private") && c.Participants.Value() == 0 {
		panic(fmt.Errorf("не получено число подписчиков для %s", channelPeer))
	}

	// Верифицирован?

	pattern = `(?s)tgme_page_title.+?` + PatternVerified

	re = regexp.Prepare("isVerified", pattern)
	if re.Match(page) {
		c.IsVerified = true
	}

	// Скам?

	patterns = []string{
		`(?i)` + PatternScamRu,
		`(?i)` + PatternScamEn,
	}

	for _, pattern := range patterns {
		re = regexp.Prepare("isScam", pattern)
		if re.Match(c.About) {
			c.IsScam = true
			c.About = ""
			break
		}
	}

	// Контактные ссылки в описании
	if c.About != "" {
		c.Contacts = check.Links(c.About, false)
		if len(c.Contacts) > 0 {
			c.Siblings = check.AdvLinks(c.Contacts, channelLink, links.Links{})
		}
	}

	chRes <- response.Response{Ok: true, Code: 200, Data: c, Error: nil}
}

var (
	messagesTriesFactor   = 10    // Среднее количество сообщений на странице
	messagesTriesSpare    = 3     // Количество запасных подгружаемых страниц
	siblingsTresholdRatio = 0.3   // Коффициент упоминания "родственной" рекламы, после которого она игнорируется
	siblingsMinMessages   = 10    // Минимальное число сообщений, при котором анализируются "родственные" ссылки
	isSkipNotextMessages  = false // Пропускать пустые сообщения?
	//timeZoneShift       = 3     // Часовой пояс
)

// Парсинг сообщений канала
func Messages(chRes chan<- response.Response, c channel.Channel, messagesCount, afterMsgId uint, isParseInfo bool) {
	defer func() {
		if err := recover(); err != nil {
			chRes <- response.Response{Ok: false, Code: 500, Error: err.(error)}
			log.Panicln("Паника парсинга сообщений:", err)
		}
	}()

	channelPeer, channelLink := format.Username(c.Username, c.Joinchat, 0)
	_, messagesLink := format.PageLink(c.Username, c.Joinchat)

	// Если нет юзернейма
	if messagesLink == "" {
		chRes <- response.Response{Ok: false, Code: 404, Error: fmt.Errorf("нет ссылки для %s", channelPeer)}
		return
	}

	// Отсечка старых сообщений
	initMessagesLink := messagesLink
	if afterMsgId != 0 {
		initMessagesLink += "?after=" + fmt.Sprint(afterMsgId)
	}

	// Получение страницы
	code, page, err := get.Page(initMessagesLink)
	if code != 200 || err != nil {
		chRes <- response.Response{Ok: false, Code: code, Error: err}
		return
	}

	initPage := page

	// Если страница невалидна
	re := regexp.Prepare("isValid", PatternIsValidMsgs)
	if !re.Match(page) {
		chRes <- response.Response{Ok: false, Code: 404, Error: fmt.Errorf("нет данных для %s", channelPeer)}
		return
	}

	// ... или пуста
	re = regexp.Prepare("isntEmpty", PatternIsntEmpty)
	if re.Match(page) {
		chRes <- response.Response{Ok: false, Code: 404, Error: fmt.Errorf("нет данных для %s", channelPeer)}
		return
	}

	c.Contacts = links.New()
	c.Siblings = links.New()

	// Парсинг описания (для родственных ссылок)

	if isParseInfo || messagesCount > 0 {
		patterns := []string{
			`(?s)` + PatternAboutBody,
			`(?s)` + PatternAboutProp,
		}

		for _, pattern := range patterns {
			re = regexp.Prepare("about", pattern)
			res := re.Find(page)

			if len(res) > 0 {
				c.About = format.SafeText(res["about"])
				break
			}
		}
	}

	// Парсинг основной информации

	if isParseInfo {

		// Название

		patterns := []string{
			PatternTitleProp,
			PatternTitleBody,
			`tgme_widget_message_owner_name.+?>(.+?>)?(?P<title>.+?)</`,
			`<title>(?P<title>.+?)\s?.\sTelegram.*?</title>`,
		}

		for _, pattern := range patterns {
			re = regexp.Prepare("title", pattern)
			res := re.Find(page)

			if len(res) > 0 {
				c.Title = format.SafeString(res["title"])
				break
			}
		}

		if c.Title == "" {
			panic(fmt.Errorf("не получено название для %s", channelPeer))
		}

		// Изображение

		patterns = []string{
			PatternImageProp,
			PatternImageBody,
			`tgme_widget_message_user_photo.+?src="(?P<image>.+?` + PatternImage + `)?"`,
		}

		for _, pattern := range patterns {
			re = regexp.Prepare("image", pattern)
			res := re.Find(page)

			if len(res) > 0 {
				c.Image = res["image"]
				break
			}
		}

		// Число подписчиков

		patterns = []string{
			`tgme_header_counter.+?>(?P<participants>` + PatternFactorNumber + `).*?</`,
			`tgme_channel_info_counters.+?"counter_value">(?P<participants>` + PatternFactorNumber + `)` +
				`</[^"]+?"counter_type">[Ss]ubscribers</`,
			`(?s)tgme_channel_info_counter[^>]+?>([^>]+?>)?(?P<participants>` + PatternFactorNumber + `)</`,
		}

		for _, pattern := range patterns {
			re = regexp.Prepare("participants", pattern)
			res := re.Find(page)

			if len(res) > 0 {
				c.Participants, err = value.New(res["participants"], false)
				break
			}
		}

		if c.Participants.Value() == 0 || err != nil {
			panic(fmt.Errorf("не получено число подписчиков для %s", channelPeer))
		}

		// Тип
		c.Kind = "channel"

		// Верифицирован?

		pattern := `tgme_(channel_info_)?header_labels.+?>` + PatternVerified

		re = regexp.Prepare("isVerified", pattern)
		if re.Match(page) {
			c.IsVerified = true
		}

		// Скам?

		patterns = []string{
			`<mark.+?(SCAM|Scam|scam).+?</mark>`,
			`(?i)` + PatternScamEn,
			`(?i)` + PatternScamRu,
		}

		wheres := []string{
			page,
			c.About,
			c.About,
		}

		for i, pattern := range patterns {
			where := wheres[i]

			re = regexp.Prepare("isScam", pattern)
			if re.Match(where) {
				c.IsScam = true
				c.About = ""
				break
			}
		}

		// Количество фото, видео, файлов и ссылок

		pattern = `tgme_channel_info_counters` +
			`(.+?"counter_value">(?P<photos>` + PatternFactorNumber + `)</[^"]+?"counter_type">[Pp]hotos</)?` +
			`(.+?"counter_value">(?P<videos>` + PatternFactorNumber + `)</[^"]+?"counter_type">[Vv]ideos</)?` +
			`(.+?"counter_value">(?P<files>` + PatternFactorNumber + `)</[^"]+?"counter_type">[Ff]iles</)?` +
			`(.+?"counter_value">(?P<links>` + PatternFactorNumber + `)</[^"]+?"counter_type">[Ll]inks</)?`

		re = regexp.Prepare("extCounts", pattern)
		res := re.Find(page)

		if len(res) > 0 {
			c.Photos, _ = value.New(res["photos"], false)
			c.Videos, _ = value.New(res["videos"], false)
			c.Files, _ = value.New(res["files"], false)
			c.Links, _ = value.New(res["links"], false)
		}
	}

	// Контактные ссылки в описании
	if c.About != "" {
		c.Contacts = check.Links(c.About, false)
		if len(c.Contacts) > 0 {
			c.Siblings = check.AdvLinks(c.Contacts, channelLink, links.Links{})
		}

		if c.Title == "" {
			c.About = ""
		}
	}

	// Парсинг сообщений

	if messagesCount == 0 {
		chRes <- response.Response{Ok: true, Code: 200, Data: c, Error: nil}
		return
	}

	var lastMessageId uint

	advsCount := make(map[string]int)
	siblingsTreshold := int(float64(messagesCount) * siblingsTresholdRatio)

	messagesTries := int(messagesCount)/messagesTriesFactor + messagesTriesSpare

	for t := 0; t < messagesTries; t++ {
		page = format.StripPage(page)
		msgs := splitMessages(page)

		// Сообщения

		for i := len(msgs) - 1; i >= 0; i-- {
			msg := msgs[i]

			nm := &message.Message{}

			nm.Links = links.New()
			nm.Advs = links.New()

			// Юзернейм и ID поста

			pattern := `(data-post="(?P<peer>.+?)/(?P<id>\d+)/?"|message_date.+?` +
				`href="((` + check.PatternHttp + `)?` + check.PatternTg + `/)?` +
				`(?P<peerExt>.+?)/(?P<idExt>\d+)/?")`

			re = regexp.Prepare("msgId", pattern)
			res := re.Find(msg)

			nm.ID, err = check.PositiveInt(res["id"])
			if nm.ID == 0 {
				nm.ID, err = check.PositiveInt(res["idExt"])
			}
			lastMessageId = nm.ID

			if len(res) == 0 || err != nil {
				panic(fmt.Errorf("не получен ID поста для %s", channelPeer))
			}

			// Сервисное?
			re = regexp.Prepare("msgIsService", `service_message`)
			if re.Match(msg[:500]) {
				continue
			}

			// Текст

			pattern = `(?s)tgme_widget_message_text[^>]+?>(<div[^>]+?>)?(?P<message>.+?)</div`

			re = regexp.Prepare("msgText", pattern)
			res = re.Find(msg)

			if len(res) > 0 {
				nm.MessageHtml = format.SafeHtml(res["message"])
				//nm.MessageText = format.SafeText(nm.MessageHtml)
			}

			// Пустое сообщение
			if nm.MessageHtml == "" && isSkipNotextMessages {
				continue
			}

			// Отредактировано?

			pattern = `tgme_widget_message_meta.+?>.*?[Ee]dited`

			re = regexp.Prepare("msgIsEdited", pattern)
			if re.Match(msg) {
				nm.IsEdited = true
			}

			// Число просмотров

			pattern = `tgme_widget_message_views.+?>(?P<views>` + PatternFactorNumber + `)</`

			re = regexp.Prepare("msgViews", pattern)
			res = re.Find(msg)

			nm.Views, err = value.New(res["views"], false)

			if len(res) == 0 || err != nil {
				panic(fmt.Errorf("не получено число просмотров для %s/%d", channelPeer, nm.ID))
			}

			// Дата публикации

			pattern = `<time.+?datetime="(?P<date>.+?)"`

			re = regexp.Prepare("msgDate", pattern)
			res = re.Find(msg)

			dateUTC, err := date.Parse(res["date"])

			if len(res) == 0 || err != nil {
				panic(fmt.Errorf("не получена дата публикации для %s/%d", channelPeer, nm.ID))
			}

			nm.Date = dateUTC.In(time.UTC)
			nm.DateLocal = nm.Date.In(time.Local)
			//nm.DateLocal = date.Local(dateUTC, timeZoneShift)

			// Репост

			pattern = `forwarded_from_name(.+?href="(?P<link>((` + check.PatternHttp + `)?` +
				check.PatternTg + `/)?.+?)(/(?P<post>\d+)/?)?")?.*?>(.+?>)?(?P<title>.+?)` +
				`</(.+?forwarded_from_author.+?>(?P<author>.+?)</)?`

			re = regexp.Prepare("msgForward", pattern)
			res = re.Find(msg)

			if len(res) > 0 {
				nm.IsForwarded = true
				nm.FwdLink = res["link"]
				nm.FwdPost, _ = check.PositiveInt(res["post"])
				nm.FwdTitle = format.SafeString(res["title"])
				nm.FwdAuthor = format.SafeString(res["author"])
			}

			// Медиа (фото, видео, документы)

			pattern = `(?s)tgme_widget_message_(?P<mediaType>photo|video|document)_wrap([^>]+?` +
				`url\s*\('(?P<mediaImage>.+?` + PatternImage + `)'\)|.+?<video[^>]+?src="(?P<mediaVideo>.+?)"|.+?` +
				`tgme_widget_message_document_title[^>]+?>(?P<mediaDocFile>.+?)</.+?` +
				`tgme_widget_message_document_extra[^>]+?>(?P<mediaDocSize>.+?)</)?`

			re = regexp.Prepare("msgMedia", pattern)
			medias := re.FindAll(msg)

			for _, media := range medias {
				switch media["mediaType"] {
				case "photo":
					nm.HasImage = true
					nm.Attachments = append(nm.Attachments, media["mediaImage"])
				case "video":
					nm.HasVideo = true
					if media["mediaVideo"] == "" {
						media["mediaVideo"] = "(media is too big)"
					}
					nm.Attachments = append(nm.Attachments, media["mediaVideo"])
				case "document":
					nm.HasDocument = true
					nm.Attachments = append(nm.Attachments, media["mediaDocFile"]+
						" ("+media["mediaDocSize"]+")")

					is := regexp.Prepare("msgDocIsImage", PatternImage+`$`)
					if is.Match(media["mediaDocFile"]) {
						nm.HasImage = true
						break
					}

					is = regexp.Prepare("msgDocIsVideo", PatternVideo+`$`)
					if is.Match(media["mediaDocFile"]) {
						nm.HasVideo = true
					}
				}
			}

			// Хэштеги

			if nm.MessageHtml != "" {
				pattern = `[^` + check.PatternARN + `/?&]` +
					`#(?P<text>[` + check.PatternARN + `][` + check.PatternARN + `_]*)`

				re = regexp.Prepare("msgHashtags", pattern)
				hashtags := re.FindAll(format.StripCode(nm.MessageHtml))

				if len(hashtags) > 0 {
					nm.Hashtags = make([]string, 0, len(hashtags))
					for _, hashtag := range hashtags {
						nm.Hashtags = append(nm.Hashtags, hashtag["text"])
					}
				}
			}

			// Ссылки
			if nm.MessageHtml != "" {
				nm.Links = check.Links(format.StripServiceTags(nm.MessageHtml), true)
			}

			// URL-кнопки

			pattern = `url_button.+?href="(?P<url>.+?)".*?>(?P<caption>.*?)</a>`

			re = regexp.Prepare("msgButtons", pattern)
			buttons := re.FindAll(msg)

			if len(buttons) > 0 {
				nm.Buttons = links.New()

				for _, button := range buttons {
					btnLink := format.Unescape(button["url"])
					btnCaption := button["caption"]

					nm.Buttons.Add(btnLink, btnCaption, 0)
					nm.Links.Add(btnLink, btnCaption, 0)
				}
			}

			// Рекламные ссылки

			if len(nm.Links) > 0 {
				nm.Advs = check.AdvLinks(nm.Links, channelLink, c.Siblings)

				// Удаление часто повторяющихся ссылок
				if int(messagesCount) >= siblingsMinMessages {
					for key, adv := range nm.Advs {
						advsCount[key]++

						if с := advsCount[key]; с >= siblingsTreshold {
							c.Siblings.Add(adv.Link, "", 0)

							nm.Advs.Remove(key)
							for m := len(c.Messages) - 1; m >= 0; m-- {
								c.Messages[m].Advs.Remove(key)
							}
						}
					}
				}
			}

			// Медиа в ссылках

			if len(nm.Links) > 0 {
				if !nm.HasImage {
					// Фото
					for _, data := range nm.Links {
						re = regexp.Prepare("msgLinkIsImage", PatternImage+`$`)
						if re.Match(data.Link) {
							nm.HasImage = true
							nm.Attachments = append(nm.Attachments, data.Link)
							break
						}
					}
				}
				if !nm.HasImage && !nm.HasVideo {
					// Видео
					for _, data := range nm.Links {
						re = regexp.Prepare("msgLinkIsVideo", PatternVideo+`$`)
						if re.Match(data.Link) {
							nm.HasVideo = true
							nm.Attachments = append(nm.Attachments, data.Link)
							break
						}
					}
				}
			}

			// Опрос или нет вложений

			if nm.MessageHtml == "" {
				re = regexp.Prepare("msgIsPoll", `tgme_widget_message_poll`)
				if re.Match(msg) {
					nm.IsPoll = true

					pattern = `(?s)` +
						`poll_question[^>]+?>(?P<question>[^<]+?)</.+?` +
						`poll_type[^>]+?>(?P<type>[^<]+?)</`

					re = regexp.Prepare("msgPollHeader", pattern)
					res = re.Find(msg)

					if len(res) > 0 {
						nm.MessageHtml += fmt.Sprintf("<b>%s</b>\n<i>%s</i>\n", res["question"], res["type"])
					}

					pattern = `(?s)` +
						`poll_option_percent[^>]+?>(?P<percent>[^<]+?)</.+?` +
						`poll_option_text[^>]+?>(?P<text>[^<]+?)</`

					re = regexp.Prepare("msgPollOptions", pattern)
					options := re.FindAll(msg)

					for _, option := range options {
						nm.MessageHtml += fmt.Sprintf("\n%s (%v)", option["text"], option["percent"])
					}

					nm.MessageHtml = format.SafeHtml(nm.MessageHtml)
					//nm.MessageText = format.SafeText(nm.MessageHtml)

				} else if !nm.HasImage && !nm.HasVideo && !nm.HasDocument {
					continue
				}
			}

			// Добавление сообщения
			c.Messages = append(c.Messages, nm)

			if len(c.Messages) >= int(messagesCount) || lastMessageId-1 <= afterMsgId || lastMessageId <= 1 {
				break
			}
		}

		if len(c.Messages) >= int(messagesCount) || lastMessageId-1 <= afterMsgId || lastMessageId <= 1 {
			break
		}

		// Подгрузка дополнительных сообщений
		moreMessagesLink := messagesLink + "?before=" + fmt.Sprint(lastMessageId)
		code, page, err = get.Page(moreMessagesLink)
		if code != 200 || err != nil ||
			format.Truncate(page, 20000) == format.Truncate(initPage, 20000) {
			chRes <- response.Response{Ok: true, Code: code, Data: c, Error: err}
			return
		}
	}

	chRes <- response.Response{Ok: true, Code: 200, Data: c, Error: nil}
}

// Разбиение страницы на сообщения
func splitMessages(page string) []string {
	res := make([]string, 0, 20)

	if page == "" {
		return res
	}

	res = strings.Split(page, "tgme_widget_message_wrap")
	if len(res) > 1 {
		res = res[1:]
	}

	return res
}

// Разбиение страницы на сообщения (regexp)
func splitMessagesRegexp(page string) []string {
	res := make([]string, 0, 20)

	if page == "" {
		return res
	}

	pattern := `(?s)message_wrap(?P<message>.+?)(tgme_widget_message_wrap|</section>)`

	re := regexp.Prepare("messages", pattern)
	msgs := re.FindAll(page)

	for _, msg := range msgs {
		res = append(res, msg["message"])
	}

	return res
}
