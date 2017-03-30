package server

import (
	"fmt"
	"time"
)

type ForumPost struct {
	Id          int       `json:"id"`
	BodyHTML    string    `json:"body_html"`
	Time        time.Time `json:"time"`
	Poster      string    `json:"poster"`
	ThreadId    int       `json:"thread_id"`
	ThreadTitle string    `json:"thread_title"`
	PageNumber  int       `json:"page_number"`
	ForumId     int       `json:"forum_id"`
	ForumName   string    `json:"forum_name"`
}

func (p *ForumPost) ActivityTime() time.Time {
	return p.Time
}

func (p *ForumPost) ActivityKey() uint32 {
	return uint32(p.Id)
}

func (p *ForumPost) PostURL() string {
	return fmt.Sprintf("https://www.pathofexile.com/forum/view-post/%v", p.Id)
}
