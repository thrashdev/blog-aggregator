package rss

import (
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type Feed struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pub_date"`
}

func FetchFeed(url string) (Feed, error) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return Feed{}, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return Feed{}, err
	}
	var rss Feed
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		fmt.Println(err)
		return Feed{}, err
	}

	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)

	newItems := []Item{}
	for _, i := range rss.Channel.Items {
		newItem := i
		newItem.Title = html.UnescapeString(newItem.Title)
		newItem.Description = html.UnescapeString(newItem.Description)
		newItems = append(newItems, newItem)
	}
	rss.Channel.Items = newItems

	return rss, nil
}
