package server

import (
	"fmt"
	"time"

	"github.com/labstack/echo"
)

type rssItem struct {
	Title       string `xml:"title"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
}

type rssChannel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
}

type rssResponse struct {
	XMLName bool       `xml:"rss"`
	Version int        `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
	Items   []rssItem  `xml:"item"`
}

func RSSHandler(db *Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		activity, _ := db.Activity(c.QueryParam("next"), 50)
		response := rssResponse{
			Version: 2,
			Channel: rssChannel{
				Title:       "GGG Tracker",
				Description: "Latest activity from Grinding Gear Games",
				Link:        AbsoluteURL(c, ""),
			},
		}
		for _, a := range activity {
			switch a.(type) {
			case *ForumPost:
				post := a.(*ForumPost)
				response.Items = append(response.Items, rssItem{
					Title:       post.Poster + " - " + post.ThreadTitle,
					Link:        post.PostURL(),
					GUID:        fmt.Sprintf("poe-forum-post-%v", post.Id),
					Description: post.BodyHTML,
					PubDate:     post.Time.Format(time.RFC1123Z),
				})
			case *RedditComment:
				comment := a.(*RedditComment)
				response.Items = append(response.Items, rssItem{
					Title:       comment.Author + " - " + comment.PostTitle,
					Link:        comment.CommentURL(),
					GUID:        "reddit-comment-" + comment.Id,
					Description: comment.BodyHTML,
					PubDate:     comment.Time.Format(time.RFC1123Z),
				})
			case *RedditPost:
				post := a.(*RedditPost)
				item := rssItem{
					Title:       post.Author + " - " + post.Title,
					Link:        "https://www.reddit.com" + post.Permalink,
					GUID:        "reddit-post-" + post.Id,
					Description: post.BodyHTML,
					PubDate:     post.Time.Format(time.RFC1123Z),
				}
				if item.Description == "" {
					item.Description = "<a href=\"" + post.URL + "\">" + post.URL + "</a>"
				}
				response.Items = append(response.Items, item)
			}
		}
		return c.XML(200, response)
	}
}
