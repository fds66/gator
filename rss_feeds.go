package main

import (
	"database/sql"
	"encoding/xml"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"

	//"html"
	"context"
	//"errors"
	"fmt"
	//"os"
	"time"

	"github.com/fds66/gator/internal/database"
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
	if len(cmd.Arguments) == 0 {
		return fmt.Errorf("no required time interval provided")
	}
	// time_between_reqs is a duration string, like 1s, 1m, 1h

	time_between_reqs := cmd.Arguments[0]
	repTime, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		fmt.Printf("Problem converting time, %v", err)
		return err
	}
	fmt.Printf("Collecting feeds every %v\n", repTime)
	ticker := time.NewTicker(repTime)

	for ; ; <-ticker.C {
		fmt.Println("calling scrapefeeds....")
		scrapeFeeds(s, context.Background())
	}

}

func handlerResetPosts(s *State, cmd Command) error {

	err := s.db.ResetPosts(context.Background())
	if err != nil {
		fmt.Printf("Problem resetting the posts database %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database posts reset")
	return nil
}

func handlerBrowse(s *State, cmd Command, user database.User) error {
	postsLimit := 2 //default limit
	var err error
	if len(cmd.Arguments) > 0 {
		postsLimit, err = strconv.Atoi(cmd.Arguments[0])
		if err != nil {
			fmt.Printf("Error converting argument to int, %v", err)
			return err
		}
	}

	postsStruct := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(postsLimit),
	}

	postsList, err := s.db.GetPostsForUser(context.Background(), postsStruct)
	if err != nil {
		fmt.Printf("Problem getting a list of posts by current user %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Posts for %s:\n", user.Name)
	//fmt.Printf("%+v", postsList)
	for i := range postsList {
		fmt.Println("")
		fmt.Printf("* Title:   		 %s\n", postsList[i].Title)
		fmt.Printf("* Link:     	 %s\n", postsList[1].Url)
		fmt.Printf("* Published on   %v\n", postsList[i].PublishedAt)
		fmt.Printf("* Description:   %s\n", postsList[i].Description.String)

	}
	return nil
}

// fetchs an rss feed from a url, return as RSSFeed pointer
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
	//printRSS(&feed)
	return &feed, nil

}

func printRSS(feed *RSSFeed) {
	fmt.Println("RSS Feed Contents:")
	fmt.Printf("%v\n", feed.Channel.Title)
	//fmt.Printf("%v\n", feed.Channel.Link)
	//fmt.Printf("%v\n", feed.Channel.Description)
	for _, item := range feed.Channel.Item {
		fmt.Printf("Title:    %v\n", item.Title)
		fmt.Printf("Link:     %v\n", item.Link)
		//fmt.Printf("%v\n", item.Description)
		fmt.Printf("Pub Date: %v\n", item.PubDate)

	}

}

// gets the body as a slice of bytes
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

// looks for the next feed to be fetched and calls the fetch function
func scrapeFeeds(s *State, ctx context.Context) error {
	nextFeed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		fmt.Printf("error returning next feed to fetch, %v\n", err)
		return err
	}
	currentTime := time.Now()
	markStruct := database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: currentTime, Valid: true},
		UpdatedAt:     currentTime,
		ID:            nextFeed.ID,
	}

	err = s.db.MarkFeedFetched(ctx, markStruct)
	if err != nil {
		fmt.Printf("error marking feed as fetched, %v\n", err)
		return err
	}

	rssFeed, err := fetchFeed(ctx, nextFeed.Url)
	if err != nil {
		fmt.Printf("error fetching feed , %v\n", err)
		return err
	}
	printRSS(rssFeed)

	//var post database.Post

	currentTime = time.Now()

	for _, item := range rssFeed.Channel.Item {
		post_id := uuid.New()
		if item.PubDate == "" {
			fmt.Printf("published date string is empty for %s", item.Title)
			continue
		}

		t, err := timeParsing(item.PubDate)
		if err != nil {
			fmt.Printf("error converting publish time ,%v", err)
		}

		postParams := database.CreatePostParams{
			ID:          post_id,
			CreatedAt:   currentTime,
			UpdatedAt:   currentTime,
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: t,
			FeedID:      nextFeed.ID,
		}
		_, err = s.db.CreatePost(ctx, postParams)
		//post, err = s.db.CreatePost(ctx, postParams)
		if err != nil {

			if !strings.Contains(err.Error(), "duplicate key") {
				fmt.Printf("error from create post, %v", err)
				fmt.Printf("post not saved %s\n", item.Title)
				return err
			} else {
				fmt.Println("duplicate")
			}
		} else {
			fmt.Printf("post saved %s\n", item.Title)
			//fmt.Printf("post saved %+v\n", post)
		}

	}

	return nil
}

// convert the publish time from the item, there may be different formats so try lots
func timeParsing(s string) (time.Time, error) {
	var t time.Time
	var err error

	layouts := []string{
		"Sat, 28 Mar 2026 20:45:54 +0000",
		time.RFC3339,          // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,      // With nanoseconds
		"2006-01-02 15:04:05", // SQL datetime
		"2006-01-02",          // Date only
		time.RFC1123Z,         // RFC 1123 with timezone
		time.RFC822Z,          // RFC 822 with timezone
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
	}
	for _, layout := range layouts {
		t, err = time.Parse(layout, s)
		if err == nil {
			//fmt.Printf("Time converted from %v to %v\n", s, t)
			return t, nil
		}
	}

	fmt.Printf("error parsing pubDate time string, %v\n", err)
	return time.Time{}, err
}
