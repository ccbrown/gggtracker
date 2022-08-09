package server

import (
	"fmt"
	"strconv"
	"time"
)

type RedditComment struct {
	Id        string    `json:"id"`
	Author    string    `json:"author"`
	BodyHTML  string    `json:"body_html"`
	PostId    string    `json:"post_id"`
	PostTitle string    `json:"post_title"`
	Time      time.Time `json:"time"`

	// Added in August 2022. Comments stored before then may not have this.
	Subreddit string `json:"subreddit"`
}

func (c *RedditComment) ActivityTime() time.Time {
	return c.Time
}

func (c *RedditComment) ActivityKey() uint32 {
	id, _ := strconv.ParseInt(c.Id, 36, 64)
	return uint32(id)
}

func (c *RedditComment) PostURL() string {
	subreddit := "pathofexile"
	if c.Subreddit != "" {
		subreddit = c.Subreddit
	}
	return fmt.Sprintf("https://www.reddit.com/r/%v/comments/%v/", subreddit, c.PostId)
}

func (c *RedditComment) CommentURL() string {
	subreddit := "pathofexile"
	if c.Subreddit != "" {
		subreddit = c.Subreddit
	}
	return fmt.Sprintf("https://www.reddit.com/r/%v/comments/%v/-/%v/?context=3", subreddit, c.PostId, c.Id)
}
