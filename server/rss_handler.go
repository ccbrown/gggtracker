package server

import (
	"fmt"
	"time"

	"github.com/labstack/echo"
)

type rssGUID struct {
	IsPermalink bool   `xml:"isPermaLink,attr"`
	GUID        string `xml:",chardata"`
}

type rssItem struct {
	Title       string  `xml:"title"`
	GUID        rssGUID `xml:"guid"`
	Description string  `xml:"description"`
	Link        string  `xml:"link"`
	PubDate     string  `xml:"pubDate"`
}

type rssAtomLink struct {
	HRef string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssChannel struct {
	Title       string      `xml:"title"`
	Description string      `xml:"description"`
	Link        string      `xml:"link"`
	AtomLink    rssAtomLink `xml:"atom:link"`
	Items       []rssItem   `xml:"item"`
}

type rssResponse struct {
	XMLName bool       `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

func RSSHandler(db Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		activity, _, err := db.Activity(LocaleForRequest(c.Request()), c.QueryParam("next"), 50, func(a Activity) bool { return true })
		if err != nil {
			return err
		}
		response := rssResponse{
			Version: "2.0",
			Atom:    "http://www.w3.org/2005/Atom",
			Channel: rssChannel{
				Title:       "GGG Tracker",
				Description: "Latest activity from Grinding Gear Games",
				Link:        AbsoluteURL(c, ""),
				AtomLink: rssAtomLink{
					HRef: AbsoluteURL(c, "/rss"),
					Rel:  "self",
					Type: "application/rss+xml",
				},
			},
		}
		for _, a := range activity {
			switch a.(type) {
			case *ForumPost:
				post := a.(*ForumPost)
				response.Channel.Items = append(response.Channel.Items, rssItem{
					Title: post.Poster + " - " + post.ThreadTitle,
					Link:  post.PostURL(),
					GUID: rssGUID{
						IsPermalink: false,
						GUID:        fmt.Sprintf("poe-forum-post-%v", post.Id),
					},
					Description: post.BodyHTML,
					PubDate:     post.Time.Format(time.RFC1123Z),
				})
			case *RedditComment:
				comment := a.(*RedditComment)
				response.Channel.Items = append(response.Channel.Items, rssItem{
					Title: comment.Author + " - " + comment.PostTitle,
					Link:  comment.CommentURL(),
					GUID: rssGUID{
						IsPermalink: false,
						GUID:        "reddit-comment-" + comment.Id,
					},
					Description: comment.BodyHTML,
					PubDate:     comment.Time.Format(time.RFC1123Z),
				})
			case *RedditPost:
				post := a.(*RedditPost)
				item := rssItem{
					Title: post.Author + " - " + post.Title,
					Link:  "https://www.reddit.com" + post.Permalink,
					GUID: rssGUID{
						IsPermalink: false,
						GUID:        "reddit-post-" + post.Id,
					},
					Description: post.BodyHTML,
					PubDate:     post.Time.Format(time.RFC1123Z),
				}
				if item.Description == "" {
					item.Description = "<a href=\"" + post.URL + "\">" + post.URL + "</a>"
				}
				response.Channel.Items = append(response.Channel.Items, item)
			}
		}
		return c.XML(200, response)
	}
}
