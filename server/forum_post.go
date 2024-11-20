package server

import (
	"encoding/json"
	"fmt"
	"time"
)

type ForumPost struct {
	Id                  int       `json:"id"`
	BodyHTML            string    `json:"body_html"`
	Time                time.Time `json:"time"`
	Poster              string    `json:"poster"`
	PosterDiscriminator int       `json:"poster_discriminator"`
	ThreadId            int       `json:"thread_id"`
	ThreadTitle         string    `json:"thread_title"`
	ForumId             int       `json:"forum_id"`
	ForumName           string    `json:"forum_name"`
}

type forumPostWithHost struct {
	Id                  int       `json:"id"`
	BodyHTML            string    `json:"body_html"`
	Time                time.Time `json:"time"`
	Poster              string    `json:"poster"`
	PosterDiscriminator int       `json:"poster_discriminator"`
	ThreadId            int       `json:"thread_id"`
	ThreadTitle         string    `json:"thread_title"`
	ForumId             int       `json:"forum_id"`
	ForumName           string    `json:"forum_name"`
	Host                string    `json:"host"`
}

func (p ForumPost) MarshalJSON() ([]byte, error) {
	return json.Marshal(&forumPostWithHost{
		Id:                  p.Id,
		BodyHTML:            p.BodyHTML,
		Time:                p.Time,
		Poster:              p.Poster,
		PosterDiscriminator: p.PosterDiscriminator,
		ThreadId:            p.ThreadId,
		ThreadTitle:         p.ThreadTitle,
		ForumId:             p.ForumId,
		ForumName:           p.ForumName,
		Host:                p.Host(),
	})
}

func (p *ForumPost) ActivityTime() time.Time {
	return p.Time
}

func (p *ForumPost) ActivityKey() uint32 {
	return uint32(p.Id)
}

func (p *ForumPost) Host() string {
	for _, l := range Locales {
		if l.ForumIds()[p.ForumId] {
			return l.ForumHost()
		}
	}
	return "www.pathofexile.com"
}

func (p *ForumPost) PostURL() string {
	return fmt.Sprintf("https://%v/forum/view-post/%v", p.Host(), p.Id)
}
