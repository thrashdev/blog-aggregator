package rss

type Feed struct {
}

type Channel struct {
	Title       string
	Link        string
	Description string
	Items       []Item
}

type Item struct {
	TItle       string
	Link        string
	Description string
	PubDate     string
}
