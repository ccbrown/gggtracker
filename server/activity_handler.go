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

func ActivityHandler(db Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		locale := LocaleForRequest(c.Request())
		filter := func(a Activity) bool {
			return true
		}
		if c.QueryParam("nohelp") == "true" {
			filter = func(a Activity) bool {
				if fp, ok := a.(*ForumPost); ok {
					if fp.ForumId == locale.HelpForumId {
						return false
					}
				}
				return true
			}
		}
		activity, next, err := db.Activity(locale, c.QueryParam("next"), 50, filter)
		if err != nil {
			return err
		}
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
