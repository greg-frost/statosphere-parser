# Statosphere Parser

Fast and reliable Go-written parser of Telegram channels via "t.me" webpages.

## Introduction

The parser can be used to get both the basic information of the channel and a list of its messages *(posts)*. This is achieved by parsing two types of web pages: `https://t.me/<username/joinchat>` and `https://t.me/s/<username>` respectively.

> Note: Here and after, the term "channel" will mean both channels, public and private, as well as chats, bots, users. However, receiving messages is only possible for public channels.

**Channel Info:**

- Title
- About *(if present)*
- Image *(avatar)* as a link
- Kind *(channel, private, chat, bot, user)*
- Number of participants *(exact or approximate)*
- Additional metrics *(photos, videos, files and links)*
- Links from the about text *(TG-links separately)*
- List of messages *(only for channels)*
- Marks "verified" and "scam"

The detailed structure of the channel can be viewed in `channel.Channel`, there are also the names of fields for JSON output, which are slightly different and can be omitted.

**Channel Message:**

- ID
- Message *(HTML and text)*
- Is a repost *(if so, by whom and where from)*
- Has the text been edited
- Is the post a poll *(if so, it is saves as text)*
- Are there any attached images/videos/documents
- List of attachments *(if present)*
- List of hashtags
- List of buttons *(URL buttons)*
- List of all links from the text
- List of Telegram links from the text *(with some exceptions)*
- Number of views *(exact or approximate)*
- Publication date *(UTC and local)*

The detailed structure of the message can be viewed in `message.Message`, including the layout of JSON fields.

The parser also uses two types of supplementary structures for values and links:

**Value** is a numeric metric that can be exact or approximate, for example, the exact number of participants of the channel: `Exact = 1234` or approximate number of message views: `Approx = 1200`, `Short = "1.2K"`.

**Links** is used to display a list of links and information on each of them, such as: `Link` - address, `Captions` - list of captions *(short ones are ignored)*, `Posts` - list of post IDs, `Pos` - position *(number in order)*, `Count` - quantity.

> Note: Not all Telegram links are of interest, but only what can be called an "advertising" or "commercial" mention, so some of the links are ignoring, namely:

- Links to the channel itself, because the authors often sign their posts in this way,
- Links to related channels *(`Siblings`)*, i.e. Telegram links specified in the channel about text,
- Frequently repeated links: for example, if out of 10 posts a link occurs in at least 3, it is added to the list of `Siblings` and no longer taken into account.

These measures are designed to make `Advs` a ready-made solution for tracking mentions, especially since all links are present in the `Links` *(by the way, URL buttons from `Buttons` are duplicated there as well)*.

## Usage

The parser can work in several modes:

### Server

This is the default mode, so to start the server, simply call the compiled file without arguments or set optional flags:

- `-mode server` or just `-server` to force start as a server,
- `-proxy=<true or false>` to enable/disable proxy *(enabled by default)*,
- `-address <server address>` and `-port <number>` to select address and port values other than "localhost" and "8080" respectively.

For example, if the compiled file is called "parser", then running in server mode might look like this:

```
./parser
```
or
```
./parser -server -address "https://example.com" -port 80 -proxy=false
```

After starting the server, you can access the pages `/`, `/info` and `/messages` without params *(then web forms will be shown for selecting the mode and parsing of information and messages)* or make GET requests using the following options:

- `channel=<username/joinchat>` or `channels=<space/comma-separated list or array>` to specify channels or you can set `test=true` flag for working with test data instead,
- `limit=<number>` and `offset=<number>` to limit and offset the selection,
- `messages=<number>` to parse messages and `exact=true` to get the exact number of participants *(only for `/messages`)*.

> Note: Test channels is a list of 150 popular IT-related channels located in the `data/channels` file, which are good for testing the parser.

The result of processing a GET request will be returned in JSON format. For the `/info` page, this will be a list of channels with main information, for `/messages` - main information *(the number of participants may not be exact)* + a list of the latest messages for each channel. The `exact=true` flag may be used to get both messages and the exact number of participants, but this will cost an additional request to the `https://t.me/<username>` page.

For example, if the server is running at `http://localhost:8080`, then getting main information about the "telegram" channel might look like this:

```
http://localhost:8080/info?channel=telegram
```

And getting the first 50 out of 150 test channels with main information *(and the exact number of participants)* and 10 messages each might look like this:

```
http://localhost:8080/messages?test=true&limit=50&messages=10&exact=true
```

**About proxy**

Proxies are needed to bypass the restrictions of the Telegram site, which allows you to make approximately 300 requests per minute from one IP address *(when the limit is exceeded, Telegram displays the page as if the channel does not exist)*.

The use of a proxy is enabled by default and, when possible, it is not enforced, giving way to regular requests. The proxy addresses located in the `data/proxies` file, but they are public and work accordingly, so for active parsing you need to enter new private proxy addresses in the same format as used in the file.

You may need to change the settings in `proxy.Enable(...)` - it set limits for requests without a proxy, with a proxy and the time for which the non-proxy "cools down", i.e. bypassing restriction of the Telegram site, so that you can return to normal work. There is also a separate mode for proxy testing, which is enabled by the `-mode proxy` flag, and testing parameters are set in `channels.TestProxy(...)`.

### Console

This parser mode is enabled using the `-mode console` or simply `-console` flags. You can then enter the following flags:

- `-channel <username/joinchat>` or `-channels <space/comma-separated list>` to specify channels or `-test` flag to use test data,
- `-limit <number>` and `-offset <number>` to limit and offset the selection,
- `-messages <number>` to parse messages and `-exact` to get the exact number of participants.

As a result, channels and messages will be parsed and displayed on the screen.

For example, if the compiled file is called "parser", then getting the first 50 out of 150 test channels with 10 messages each in console mode might look like this:

```
./parser -console -test -messages 10 -limit 50
```

### Package

In the case of using the parser as a package, you need to import the module *(package `parse`)* and write code like this:

```
// Initialization
channels := parse.NewChannels()

// Preparing

channels.Add("username") // by one,
channels.PrepareFromString("username1, username2 username3") // from string
channels.PrepareFromFile("data/channels") // or from a file

// Limiting
channels.Limit(0, 50)

// Parsing

messages := 10 // required number of messages
isExact := true // get the exact number of participants
_, errs := channels.Parse(context.Background(), isExact, messages)

// Using the results

if channel, ok := channels.Find("username"); ok { // search for one
	fmt.Println(channel.Title);
	fmt.Println(channel.About);
	fmt.Println(channel.Kind);
	// ...

	// (full list of available fields
	// located in channel.Channel)
}
for _, channel := range channels.Channels.Channels { // or range all
	// ...
}

// You can also just print the results to the console
channels.Print(true) // false - info only, true - with messages

// Or output to JSON
// (will need to import the "response" package)

channels.RemoveUnparsed() // remove channels without info
channels.RemoveUnmessaged() // remove channels without messages

json := response.New(
	channels.Channels.Channels, channels.Count(),
	errs, time.Since(start),
)
result, _ := json.EncodeJSON()
response.PrintJSON(res, result)
```

## P.S.

I go to the Go language just recently, and wrote this parser rather to practice and create a representative demo project that would use everything I learned, but if this code helps someone, I will be glad. And in general, it was interesting to participate a little in Open Source. By the way, there is also an [Russian version](README_ru.md) of this manual. Good luck!