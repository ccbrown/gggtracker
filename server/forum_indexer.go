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

type ForumAccount struct {
	Username      string
	Discriminator int
}

func (indexer *ForumIndexer) run() {
	log.Info("starting forum indexer")

	accounts := []ForumAccount{
		{Username: "Chris"},
		{Username: "Jonathan"},
		{Username: "Mark_GGG"},
		{Username: "Rory"},
		{Username: "Rhys"},
		{Username: "Joel_GGG"},
		{Username: "Ari"},
		{Username: "Thomas"},
		{Username: "Support"},
		{Username: "Jess_GGG"},
		{Username: "Robbie_GGG"},
		{Username: "GGG_Neon"},
		{Username: "Jason_GGG"},
		{Username: "Henry_GGG"},
		{Username: "Michael_GGG"},
		{Username: "Bex_GGG"},
		{Username: "Cagan_GGG"},
		{Username: "Kieren_GGG"},
		{Username: "Yeran_GGG"},
		{Username: "Gary_GGG"},
		{Username: "Dan_GGG"},
		{Username: "Jared_GGG"},
		{Username: "Brian_GGG"},
		{Username: "RobbieL_GGG"},
		{Username: "Arthur_GGG"},
		{Username: "NickK_GGG"},
		{Username: "Felipe_GGG"},
		{Username: "Alex_GGG"},
		{Username: "Alexcc_GGG"},
		{Username: "CJ_GGG"},
		{Username: "Eben_GGG"},
		{Username: "Emma_GGG"},
		{Username: "Ethan_GGG"},
		{Username: "Fitzy_GGG"},
		{Username: "Hartlin_GGG"},
		{Username: "Jake_GGG"},
		{Username: "Melissa_GGG"},
		{Username: "MikeP_GGG"},
		{Username: "Novynn"},
		{Username: "Rob_GGG"},
		{Username: "Roman_GGG"},
		{Username: "Tom_GGG"},
		{Username: "Natalia_GGG"},
		{Username: "Jeff_GGG"},
		{Username: "Lu_GGG"},
		{Username: "JuliaS_GGG"},
		{Username: "Alexander_GGG"},
		{Username: "SamC_GGG"},
		{Username: "AndrewE_GGG"},
		{Username: "Kyle_GGG"},
		{Username: "Stacey_GGG"},
		{Username: "Jatin_GGG"},
		{Username: "Community_Team"},
		{Username: "Nick_GGG"},
		{Username: "Guy_GGG"},
		{Username: "Ben_GGG"},
		{Username: "BenH_GGG"},
		{Username: "Nav_GGG"},
		{Username: "Will_GGG"},
		{Username: "Scott_GGG"},
		{Username: "JC_GGG"},
		{Username: "Dylan_GGG"},
		{Username: "Chulainn_GGG"},
		{Username: "Vash_GGG"},
		{Username: "Cameron_GGG"},
		{Username: "Jacob_GGG"},
		{Username: "Jenn_GGG"},
		{Username: "CoryA_GGG"},
		{Username: "Sian_GGG"},
		{Username: "Drew_GGG"},
		{Username: "Lisa_GGG"},
		{Username: "ThomasK_GGG"},
		{Username: "Whai_GGG"},
		{Username: "Scopey"},
		{Username: "Adam_GGG"},
		{Username: "Nichelle_GGG"},
		{Username: "Markus_GGG"},
		{Username: "Jarod_GGG"},
		{Username: "Joel_GGG", Discriminator: 1496},
		{Username: "Vinky_GGG"},
		{Username: "Edmund_GGG", Discriminator: 4844},
		{Username: "Clint"},
		{Username: "LeightonJ_GGG"},
		{Username: "Tai_GGG"},
		{Username: "ShaunB_GGG"},
		{Username: "Ayelen_GGG"},
		{Username: "Timothy_GGG"},
		{Username: "BenMH_GGG"},
		{Username: "Ian_GGG"},
		{Username: "EthanH_GGG"},
		{Username: "Yrone_GGG", Discriminator: 9576},
		{Username: "Sam_GGG", Discriminator: 2420},
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
					if errors.Is(err, ErrForumMaintenance) {
						log.Info("forum is under maintenance")
						time.Sleep(30 * time.Second)
					} else {
						log.WithError(err).Error("error indexing forum account: " + account.Username)
					}
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

var ErrForumMaintenance = errors.New("forum is in maintenance")

func ScrapeForumPosts(doc *goquery.Document, poster ForumAccount, timezone *time.Location) ([]*ForumPost, error) {
	posts := []*ForumPost(nil)

	err := error(nil)

	if doc.Find(".forumPostListTable").Length() == 0 {
		err = errors.New("forum post list not found")
		if topBar := doc.Find(".topBar"); topBar.Length() == 1 && topBar.Text() == "Down For Maintenance" {
			err = ErrForumMaintenance
		}
		return nil, err
	}

	doc.Find(".forumPostListTable > tbody > tr").EachWithBreak(func(i int, sel *goquery.Selection) bool {
		post := &ForumPost{
			Poster:              poster.Username,
			PosterDiscriminator: poster.Discriminator,
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

func (indexer *ForumIndexer) forumPosts(poster ForumAccount, page int, timezone *time.Location) ([]*ForumPost, error) {
	doc, err := indexer.requestDocument(fmt.Sprintf("/account/view-posts/%v-%04d/page/%v", poster.Username, poster.Discriminator, page))
	if err != nil {
		return nil, err
	}
	posts, err := ScrapeForumPosts(doc, poster, timezone)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (indexer *ForumIndexer) index(poster ForumAccount, timezone *time.Location) error {
	logger := log.WithFields(log.Fields{
		"poster": poster,
	})

	pageCutoff := time.Now().Add(-12 * time.Hour)
	cutoff := time.Now().Add(-14 * 24 * time.Hour)
	activity := []Activity(nil)

	for page := 1; ; page++ {
		posts, err := indexer.forumPosts(poster, page, timezone)
		if err != nil {
			return fmt.Errorf("error getting forum posts: %w", err)
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
