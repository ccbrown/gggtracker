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

	log "github.com/sirupsen/logrus"
)

const UserAgent = "gggtracker.com (github.com/ccbrown/gggtracker)"

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
		"Chris", "Jonathan", "Erik", "Mark_GGG", "Samantha", "Rory", "Rhys", "Andrew_GGG",
		"Damien_GGG", "Joel_GGG", "Ari", "Thomas", "BrianWeissman", "Edwin_GGG", "Support", "Dylan",
		"MaxS", "Ammon_GGG", "Jess_GGG", "Robbie_GGG", "GGG_Neon", "Jason_GGG", "Henry_GGG",
		"Michael_GGG", "Bex_GGG", "Cagan_GGG", "Daniel_GGG", "Kieren_GGG", "Yeran_GGG", "Gary_GGG",
		"Dan_GGG", "Jared_GGG", "Brian_GGG", "RobbieL_GGG", "Arthur_GGG", "NickK_GGG", "Felipe_GGG",
		"Alex_GGG", "Alexcc_GGG", "Andy", "CJ_GGG", "Eben_GGG", "Emma_GGG", "Ethan_GGG",
		"Fitzy_GGG", "Hartlin_GGG", "Jake_GGG", "Lionel_GGG", "Melissa_GGG", "MikeP_GGG", "Novynn",
		"Rachel_GGG", "Rob_GGG", "Roman_GGG", "Sarah_GGG", "SarahB_GGG", "Tom_GGG", "Natalia_GGG",
		"Jeff_GGG", "Lu_GGG", "JuliaS_GGG", "Alexander_GGG", "SamC_GGG", "AndrewE_GGG", "Kyle_GGG",
		"Stacey_GGG", "Jatin_GGG", "Yolandi_GGG", "Community_Team", "Dominic_GGG", "Nick_GGG",
		"Guy_GGG", "Ben_GGG", "BenH_GGG", "Nav_GGG", "Will_GGG", "Scott_GGG", "JC_GGG", "Dylan_GGG",
		"Chulainn_GGG",
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

	for {
		for _, locale := range Locales {
			select {
			case <-indexer.closeSignal:
				return
			default:
				logger := log.WithField("host", locale.ForumHost())
				if err := locale.RefreshForumIds(); err != nil {
					logger.WithError(err).Error("error refreshing forum ids")
				} else {
					logger.Info("refreshed forum ids")
				}
				time.Sleep(time.Second)
			}
		}
		for _, account := range accounts {
			select {
			case <-indexer.closeSignal:
				return
			default:
				if err := indexer.index(account, timezone); err != nil {
					log.WithError(err).Error("error indexing forum account: " + account)
				}
				time.Sleep(time.Second)
			}
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

	req, err := http.NewRequest("GET", urlString, nil)
	req.Header.Set("User-Agent", UserAgent)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return goquery.NewDocumentFromReader(resp.Body)
}

var postURLExpression = regexp.MustCompile("^/forum/view-post/([0-9]+)")
var threadURLExpression = regexp.MustCompile("^/forum/view-thread/([0-9]+)")
var forumURLExpression = regexp.MustCompile("^/forum/view-forum/([0-9]+)")

func ScrapeForumPosts(doc *goquery.Document, timezone *time.Location) ([]*ForumPost, error) {
	posts := []*ForumPost(nil)

	err := error(nil)

	doc.Find(".forumPostListTable > tbody > tr").EachWithBreak(func(i int, sel *goquery.Selection) bool {
		post := &ForumPost{
			Poster: sel.Find(".post_by_account").Text(),
		}

		body, err := sel.Find(".content").Html()
		if err != nil {
			return false
		}
		post.BodyHTML = body

		timeText := sel.Find(".post_date").Text()

		if post.Time, err = time.ParseInLocation("Jan _2, 2006, 3:04:05 PM", timeText, timezone); err != nil {
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

func (indexer *ForumIndexer) forumPosts(poster string, page int, timezone *time.Location) ([]*ForumPost, error) {
	doc, err := indexer.requestDocument(fmt.Sprintf("/account/view-posts/%v/page/%v", poster, page))
	if err != nil {
		return nil, err
	}
	posts, err := ScrapeForumPosts(doc, timezone)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (indexer *ForumIndexer) index(poster string, timezone *time.Location) error {
	logger := log.WithFields(log.Fields{
		"poster": poster,
	})

	pageCutoff := time.Now().Add(-12 * time.Hour)
	cutoff := time.Now().Add(-14 * 24 * time.Hour)
	activity := []Activity(nil)

	for page := 1; ; page++ {
		posts, err := indexer.forumPosts(poster, page, timezone)
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
	doc, err := indexer.requestDocument("/my-account/preferences")
	if err != nil {
		return nil, err
	}
	return ScrapeForumTimezone(doc)
}
