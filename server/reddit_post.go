package server

import (
	"strconv"
	"time"
)

type RedditPost struct {
	Id        string    `json:"id"`
	Author    string    `json:"author"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	BodyHTML  string    `json:"body_html,omitempty"`
	Permalink string    `json:"permalink"`
	Time      time.Time `json:"time"`

	// Added in August 2022. Comments stored before then may not have this.
	Subreddit string `json:"subreddit"`
}

func (p *RedditPost) ActivityTime() time.Time {
	return p.Time
}

func (p *RedditPost) ActivityKey() uint32 {
	id, _ := strconv.ParseInt(p.Id, 36, 64)
	return uint32(id)
}
