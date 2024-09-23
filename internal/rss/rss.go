package rss

import (
	"encoding/xml"
	"fmt"
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

	return rss, nil
}
