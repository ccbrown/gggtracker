package server

import (
	"github.com/labstack/echo"
)

type jsonResponseActivity struct {
	Type string   `json:"type"`
	Data Activity `json:"data"`
}

type jsonResponse struct {
	Activity []*jsonResponseActivity `json:"activity"`
	Next     string                  `json:"next"`
}

func ActivityHandler(db *Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		activity, next := db.Activity(c.QueryParam("next"), 50, LocaleForRequest(c.Request()).ActivityFilter)
		response := jsonResponse{
			Next: next,
		}
		for _, a := range activity {
			t := ""
			switch a.(type) {
			case *ForumPost:
				t = "forum_post"
			case *RedditComment:
				t = "reddit_comment"
			case *RedditPost:
				t = "reddit_post"
			}
			response.Activity = append(response.Activity, &jsonResponseActivity{
				Type: t,
				Data: a,
			})
		}
		return c.JSON(200, response)
	}
}
