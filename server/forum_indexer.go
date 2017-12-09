package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	log "github.com/Sirupsen/logrus"
)

type ForumIndexer struct {
	configuration ForumIndexerConfiguration
	closeSignal   chan bool
}

type ForumIndexerConfiguration struct {
	Database *Database
	Session  string
}

func NewForumIndexer(configuration ForumIndexerConfiguration) (*ForumIndexer, error) {
	ret := &ForumIndexer{
		configuration: configuration,
		closeSignal:   make(chan bool),
	}
	go ret.run()
	return ret, nil
}

func (indexer *ForumIndexer) Close() {
	indexer.closeSignal <- true
}

func (indexer *ForumIndexer) run() {
	log.Info("starting forum indexer")

	posters := []string{
		"Chris", "Jonathan", "Erik", "Mark_GGG", "Samantha", "Rory", "Rhys", "Qarl", "Andrew_GGG",
		"Damien_GGG", "Joel_GGG", "Ari", "Thomas", "BrianWeissman", "Edwin_GGG", "Support", "Dylan",
		"MaxS", "Ammon_GGG", "Jess_GGG", "Robbie_GGG", "GGG_Neon", "Jason_GGG", "Henry_GGG",
		"Michael_GGG", "Bex_GGG", "Cagan_GGG", "Daniel_GGG", "Kieren_GGG", "Yeran_GGG", "Gary_GGG",
		"Dan_GGG", "Jared_GGG", "Brian_GGG", "RobbieL_GGG", "Arthur_GGG", "NickK_GGG", "Felipe_GGG",
		"Alex_GGG", "Alexcc_GGG", "Andy", "CJ_GGG", "Eben_GGG", "Emma_GGG", "Ethan_GGG",
		"Fitzy_GGG", "Hartlin_GGG", "Jake_GGG", "Lionel_GGG", "Melissa_GGG", "MikeP_GGG", "Novynn",
		"Rachel_GGG", "Rob_GGG", "Roman_GGG", "Sarah_GGG", "SarahB_GGG", "Tom_GGG", "Natalia_GGG",
	}
	next := 0

	timezone := (*time.Location)(nil)

	for {
		select {
		case <-indexer.closeSignal:
			return
		default:
			if timezone == nil {
				tz, err := indexer.sessionTimezone()
				if err != nil {
					log.WithError(err).Error("error getting forum timezone")
				} else {
					timezone = tz
					log.WithFields(log.Fields{
						"timezone": timezone,
					}).Info("forum timezone obtained")
				}
			} else {
				indexer.index(posters[next], timezone)
				next += 1
				if next >= len(posters) {
					next = 0
				}
			}
			time.Sleep(time.Second)
		}
	}
}

func (indexer *ForumIndexer) requestDocument(resource string) (*goquery.Document, error) {
	urlString := fmt.Sprintf("https://www.pathofexile.com/%v", strings.TrimPrefix(resource, "/"))
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse(urlString)
	jar.SetCookies(u, []*http.Cookie{
		{
			Name:   "POESESSID",
			Value:  indexer.configuration.Session,
			Path:   "/",
			Domain: ".pathofexile.com",
		},
	})
	client := http.Client{
		Jar:     jar,
		Timeout: time.Second * 10,
	}
	resp, err := client.Get(urlString)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return goquery.NewDocumentFromReader(resp.Body)
}

var postURLExpression = regexp.MustCompile("^/forum/view-thread/([0-9]+)/page/([0-9]+)#p([0-9]+)")
var threadURLExpression = regexp.MustCompile("^/forum/view-thread/([0-9]+)")
var forumURLExpression = regexp.MustCompile("^/forum/view-forum/([0-9]+)")

func ScrapeForumPosts(doc *goquery.Document, timezone *time.Location) ([]*ForumPost, error) {
	posts := []*ForumPost(nil)

	err := error(nil)

	doc.Find(".forumPostListTable tr").EachWithBreak(func(i int, sel *goquery.Selection) bool {
		post := &ForumPost{
			Poster: sel.Find(".post_by_account").Text(),
		}

		body, err := sel.Find(".content").Html()
		if err != nil {
			return false
		}
		post.BodyHTML = body

		text := sel.Find(".post_date").Text()
		t, err := time.ParseInLocation("Jan _2, 2006 3:04:05 PM", text, timezone)
		if err != nil {
			return false
		}
		post.Time = t

		sel.Find("a").Each(func(i int, sel *goquery.Selection) {
			href := sel.AttrOr("href", "")
			if match := postURLExpression.FindStringSubmatch(href); match != nil {
				n, _ := strconv.Atoi(match[1])
				post.ThreadId = n
				n, _ = strconv.Atoi(match[2])
				post.PageNumber = n
				n, _ = strconv.Atoi(match[3])
				post.Id = n
			} else if match := threadURLExpression.FindStringSubmatch(href); match != nil {
				post.ThreadTitle = sel.Text()
			} else if match := forumURLExpression.FindStringSubmatch(href); match != nil {
				n, _ := strconv.Atoi(match[1])
				post.ForumId = n
				post.ForumName = sel.Text()
			}
		})

		posts = append(posts, post)
		return true
	})

	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (indexer *ForumIndexer) forumPosts(poster string, page int, timezone *time.Location) ([]*ForumPost, error) {
	doc, err := indexer.requestDocument(fmt.Sprintf("/account/view-posts/%v/page/%v", poster, page))
	if err != nil {
		return nil, err
	}
	return ScrapeForumPosts(doc, timezone)
}

func (indexer *ForumIndexer) index(poster string, timezone *time.Location) {
	logger := log.WithFields(log.Fields{
		"poster": poster,
	})

	cutoff := time.Now().Add(time.Hour * -12)
	activity := []Activity(nil)

	for page := 1; ; page++ {
		posts, err := indexer.forumPosts(poster, page, timezone)
		if err != nil {
			logger.WithError(err).Error("error requesting forum posts")
		}

		done := len(posts) == 0
		for _, post := range posts {
			if post.Time.Before(cutoff) {
				done = true
			}
			activity = append(activity, post)
		}

		logger.WithFields(log.Fields{
			"count": len(posts),
		}).Info("received forum posts")

		if done {
			break
		}
		time.Sleep(time.Second)
	}

	if len(activity) > 0 {
		indexer.configuration.Database.AddActivity(activity)
	}
}

func ScrapeForumTimezone(doc *goquery.Document) (*time.Location, error) {
	sel := doc.Find(`select[name="preferences[timezone]"] option[selected]`)
	if sel == nil || sel.AttrOr("value", "") == "" {
		return nil, errors.New("unable to find timezone selection")
	}
	return time.LoadLocation(sel.AttrOr("value", ""))
}

func (indexer *ForumIndexer) sessionTimezone() (*time.Location, error) {
	doc, err := indexer.requestDocument("/my-account/preferences")
	if err != nil {
		return nil, err
	}
	return ScrapeForumTimezone(doc)
}
