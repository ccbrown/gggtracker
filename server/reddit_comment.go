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
}

func (c *RedditComment) ActivityTime() time.Time {
	return c.Time
}

func (c *RedditComment) ActivityKey() uint32 {
	id, _ := strconv.ParseInt(c.Id, 36, 64)
	return uint32(id)
}

func (c *RedditComment) PostURL() string {
	return fmt.Sprintf("https://www.reddit.com/r/pathofexile/comments/%v/", c.PostId)
}

func (c *RedditComment) CommentURL() string {
	return fmt.Sprintf("https://www.reddit.com/r/pathofexile/comments/%v/-/%v/?context=3", c.PostId, c.Id)
}
