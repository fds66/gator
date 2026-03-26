package main

import (
	"encoding/xml"
	"html"
	"io"
	"net/http"

	//"html"
	"context"
	//"errors"
	"fmt"
	//"os"
	"time"
	//"github.com/google/uuid"
	//"github.com/fds66/gator/internal/config"
	//"github.com/fds66/gator/internal/database"
)

type Client struct {
	httpClient http.Client
}

func NewClient(timeout time.Duration) Client {
	return Client{
		httpClient: http.Client{
			Timeout: timeout,
		},
	}

}

/*
<rss xmlns:atom="http://www.w3.org/2005/Atom" version="2.0">
<channel>

	<title>RSS Feed Example</title>
	<link>https://www.example.com</link>
	<description>This is an example RSS feed</description>
	<item>
	  <title>First Article</title>
	  <link>https://www.example.com/article1</link>
	  <description>This is the content of the first article.</description>
	  <pubDate>Mon, 06 Sep 2021 12:00:00 GMT</pubDate>
	</item>
	<item>
	  <title>Second Article</title>
	  <link>https://www.example.com/article2</link>
	  <description>Here's the content of the second article.</description>
	  <pubDate>Tue, 07 Sep 2021 14:30:00 GMT</pubDate>
	</item>

</channel>
</rss>
*/
type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func handlerAgg(s *State, cmd Command) error {
	feedURL := "https://www.wagslane.dev/index.xml"
	RSScontent, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		fmt.Printf("Problem accessing RSS feed, %v", err)
	}
	fmt.Printf("RSS Feed Struct: %+v", RSScontent)
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	//feedURL = "https://www.wagslane.dev/index.xml"
	feed := RSSFeed{}
	body, err := getRSSbody(ctx, feedURL)
	if err != nil {
		return &feed, err
	}
	//fmt.Println(string(body)) //if I want to check the body while debugging
	//xml.Unmarshal works same as json.Unmarshal
	if err := xml.Unmarshal(body, &feed); err != nil {
		return &feed, err
	}
	// run Title and Description fields of both channel and items through html.UnescapeString
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i, item := range feed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		feed.Channel.Item[i] = item
	}

	return &feed, nil

}

func getRSSbody(ctx context.Context, feedURL string) ([]byte, error) {
	//takes in url //returns the body

	// create the client
	RSSClient := NewClient(5 * time.Second)
	// form the request
	var body []byte
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return body, err
	}
	// request.header.set User-Agent to gator
	req.Header.Set("User-Agent", "gator")
	results, err := RSSClient.httpClient.Do(req)
	if err != nil {
		return body, err
	}
	defer results.Body.Close()
	body, err = io.ReadAll(results.Body)
	if err != nil {
		return body, err
	}
	return body, nil
}
