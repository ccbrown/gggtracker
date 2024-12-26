package server

import (
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
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
		activity, next, err := fetchActivity(db, LocaleForRequest(c.Request()), c.QueryParam("next"), c.QueryParam("nohelp") == "true")
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

const MinPageSize = 50
const DbRequestSize = 50
const MaxDbRequests = 10

func fetchActivity(db Database, locale *Locale, start string, nohelp bool) ([]Activity, string, error) {
	activity := []Activity{}
	next := start
	pred := func(a Activity) bool {
		return true
	}
	if nohelp {
		pred = func(a Activity) bool {
			if fp, ok := a.(*ForumPost); ok {
				if fp.ForumId == locale.HelpForumId {
					return false
				}
			}
			return true
		}
	}
	for i := 0; i < MaxDbRequests && len(activity) < MinPageSize; i++ {
		as, n, err := db.Activity(locale, next, DbRequestSize)
		if err != nil {
			return nil, "", err
		}
		next = n
		if len(as) == 0 && n == "" {
			log.Debug("end of activity db")
			break
		}
		filtered := 0
		for _, a := range as {
			if pred(a) {
				activity = append(activity, a)
			} else {
				filtered++
			}
		}
		log.WithFields(log.Fields{
			"count":    len(as),
			"buffered": len(activity),
			"filtered": filtered,
			"next":     next,
		}).Debug("processed activity batch")
	}
	return activity, next, nil
}
