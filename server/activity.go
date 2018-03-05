package server

import (
	"net/http"
	"strings"
	"time"
)

type Activity interface {
	ActivityTime() time.Time
	ActivityKey() uint32
}

func ActivityFilterForRequest(r *http.Request) func(Activity) bool {
	subdomain := ""
	if r.Host != "" {
		subdomain = strings.Split(r.Host, ".")[0]
	}

	includeReddit := false
	includeForumHost := map[string]bool{}

	switch subdomain {
	case "br":
		includeForumHost["br.pathofexile.com"] = true
	case "ru":
		includeForumHost["ru.pathofexile.com"] = true
	case "th":
		includeForumHost["th.pathofexile.com"] = true
	case "de":
		includeForumHost["de.pathofexile.com"] = true
	case "fr":
		includeForumHost["fr.pathofexile.com"] = true
	case "es":
		includeForumHost["es.pathofexile.com"] = true
	default:
		includeReddit = true
		includeForumHost["www.pathofexile.com"] = true
	}

	return func(a Activity) bool {
		switch a := a.(type) {
		case *ForumPost:
			return includeForumHost[a.Host]
		case *RedditComment:
			return includeReddit
		case *RedditPost:
			return includeReddit
		}
		return false
	}
}
