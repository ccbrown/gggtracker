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
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"

	log "github.com/sirupsen/logrus"
)

type ForumIndexer struct {
	configuration ForumIndexerConfiguration
	closeSignal   chan struct{}
}

type ForumIndexerConfiguration struct {
	Database Database
	Session  string
}

func NewForumIndexer(configuration ForumIndexerConfiguration) (*ForumIndexer, error) {
	ret := &ForumIndexer{
		configuration: configuration,
		closeSignal:   make(chan struct{}),
	}
	go ret.run()
	return ret, nil
}

func (indexer *ForumIndexer) Close() {
	close(indexer.closeSignal)
}

func (indexer *ForumIndexer) run() {
	log.Info("starting forum indexer")

	accounts := []string{
		"Chris", "Jonathan", "Erik", "Mark_GGG", "Samantha", "Rory", "Rhys", "Qarl", "Andrew_GGG",
		"Damien_GGG", "Joel_GGG", "Ari", "Thomas", "BrianWeissman", "Edwin_GGG", "Support", "Dylan",
		"MaxS", "Ammon_GGG", "Jess_GGG", "Robbie_GGG", "GGG_Neon", "Jason_GGG", "Henry_GGG",
		"Michael_GGG", "Bex_GGG", "Cagan_GGG", "Daniel_GGG", "Kieren_GGG", "Yeran_GGG", "Gary_GGG",
		"Dan_GGG", "Jared_GGG", "Brian_GGG", "RobbieL_GGG", "Arthur_GGG", "NickK_GGG", "Felipe_GGG",
		"Alex_GGG", "Alexcc_GGG", "Andy", "CJ_GGG", "Eben_GGG", "Emma_GGG", "Ethan_GGG",
		"Fitzy_GGG", "Hartlin_GGG", "Jake_GGG", "Lionel_GGG", "Melissa_GGG", "MikeP_GGG", "Novynn",
		"Rachel_GGG", "Rob_GGG", "Roman_GGG", "Sarah_GGG", "SarahB_GGG", "Tom_GGG", "Natalia_GGG",
		"Jeff_GGG", "Lu_GGG", "JuliaS_GGG", "Alexander_GGG", "SamC_GGG", "AndrewE_GGG", "Kyle_GGG",
	}

	timezone := (*time.Location)(nil)

	for timezone == nil {
		select {
		case <-indexer.closeSignal:
			return
		default:
			if tz, err := indexer.sessionTimezone(); err != nil {
				log.WithError(err).Error("error getting forum timezone")
			} else {
				timezone = tz
				log.WithField("timezone", timezone).Info("forum timezone obtained")
				break
			}
			time.Sleep(time.Second)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(Locales))

	for _, l := range Locales {
		l := l
		go func() {
			for {
				for _, account := range accounts {
					select {
					case <-indexer.closeSignal:
						return
					default:
						if err := indexer.index(l, account, timezone); err != nil {
							log.WithError(err).Error("error indexing forum account: " + account)
						}
						time.Sleep(time.Second)
					}
				}
			}
		}()
	}

	wg.Wait()
}

func (indexer *ForumIndexer) requestDocument(host, resource string) (*goquery.Document, error) {
	urlString := fmt.Sprintf("https://%v/%v", host, strings.TrimPrefix(resource, "/"))
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

var postURLExpression = regexp.MustCompile("^/forum/view-post/([0-9]+)")
var threadURLExpression = regexp.MustCompile("^/forum/view-thread/([0-9]+)")
var forumURLExpression = regexp.MustCompile("^/forum/view-forum/([0-9]+)")

func ScrapeForumPosts(doc *goquery.Document, locale *Locale, timezone *time.Location) ([]*ForumPost, error) {
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

		timeText := sel.Find(".post_date").Text()

		if post.Time, err = locale.ParseTime(timeText, timezone); err != nil {
			log.WithField("text", timeText).Error("unable to parse time")
			return false
		}

		sel.Find("a").Each(func(i int, sel *goquery.Selection) {
			href := sel.AttrOr("href", "")
			if match := postURLExpression.FindStringSubmatch(href); match != nil {
				n, _ := strconv.Atoi(match[1])
				post.Id = n
			} else if match := threadURLExpression.FindStringSubmatch(href); match != nil {
				n, _ := strconv.Atoi(match[1])
				post.ThreadId = n
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

func (indexer *ForumIndexer) forumPosts(locale *Locale, poster string, page int, timezone *time.Location) ([]*ForumPost, error) {
	doc, err := indexer.requestDocument(locale.ForumHost(), fmt.Sprintf("/account/view-posts/%v/page/%v", poster, page))
	if err != nil {
		return nil, err
	}
	posts, err := ScrapeForumPosts(doc, locale, timezone)
	if err != nil {
		return nil, err
	}
	for _, post := range posts {
		post.Host = locale.ForumHost()
	}
	return posts, nil
}

func (indexer *ForumIndexer) index(locale *Locale, poster string, timezone *time.Location) error {
	logger := log.WithFields(log.Fields{
		"host":   locale.ForumHost(),
		"poster": poster,
	})

	pageCutoff := time.Now().Add(-12 * time.Hour)
	cutoff := time.Now().Add(-14 * 24 * time.Hour)
	activity := []Activity(nil)

	for page := 1; ; page++ {
		posts, err := indexer.forumPosts(locale, poster, page, timezone)
		if err != nil {
			logger.WithError(err).Error("error requesting forum posts")
		}

		done := len(posts) == 0
		for _, post := range posts {
			if post.Time.Before(pageCutoff) {
				done = true
			}
			if post.Time.Before(cutoff) {
				break
			}
			activity = append(activity, post)
		}

		logger.WithField("count", len(activity)).Info("received forum posts")

		if done {
			break
		}
		time.Sleep(time.Second)
	}

	if len(activity) == 0 {
		return nil
	}

	return indexer.configuration.Database.AddActivity(activity)
}

func ScrapeForumTimezone(doc *goquery.Document) (*time.Location, error) {
	sel := doc.Find(`select[name="preferences[timezone]"] option[selected]`)
	if sel == nil || sel.AttrOr("value", "") == "" {
		return nil, errors.New("unable to find timezone selection")
	}
	return time.LoadLocation(sel.AttrOr("value", ""))
}

func (indexer *ForumIndexer) sessionTimezone() (*time.Location, error) {
	doc, err := indexer.requestDocument("www.pathofexile.com", "/my-account/preferences")
	if err != nil {
		return nil, err
	}
	return ScrapeForumTimezone(doc)
}
