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
	Database *Database
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

	hosts := []string{
		"www.pathofexile.com",
		"br.pathofexile.com",
		"ru.pathofexile.com",
		"th.pathofexile.com",
		"de.pathofexile.com",
		"fr.pathofexile.com",
		"es.pathofexile.com",
	}

	accounts := []string{
		"Chris", "Jonathan", "Erik", "Mark_GGG", "Samantha", "Rory", "Rhys", "Qarl", "Andrew_GGG",
		"Damien_GGG", "Joel_GGG", "Ari", "Thomas", "BrianWeissman", "Edwin_GGG", "Support", "Dylan",
		"MaxS", "Ammon_GGG", "Jess_GGG", "Robbie_GGG", "GGG_Neon", "Jason_GGG", "Henry_GGG",
		"Michael_GGG", "Bex_GGG", "Cagan_GGG", "Daniel_GGG", "Kieren_GGG", "Yeran_GGG", "Gary_GGG",
		"Dan_GGG", "Jared_GGG", "Brian_GGG", "RobbieL_GGG", "Arthur_GGG", "NickK_GGG", "Felipe_GGG",
		"Alex_GGG", "Alexcc_GGG", "Andy", "CJ_GGG", "Eben_GGG", "Emma_GGG", "Ethan_GGG",
		"Fitzy_GGG", "Hartlin_GGG", "Jake_GGG", "Lionel_GGG", "Melissa_GGG", "MikeP_GGG", "Novynn",
		"Rachel_GGG", "Rob_GGG", "Roman_GGG", "Sarah_GGG", "SarahB_GGG", "Tom_GGG", "Natalia_GGG",
		"Jeff_GGG",
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
	wg.Add(len(hosts))

	for _, host := range hosts {
		host := host
		go func() {
			for {
				for _, account := range accounts {
					select {
					case <-indexer.closeSignal:
						return
					default:
						indexer.index(host, account, timezone)
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

var postURLExpression = regexp.MustCompile("^/forum/view-thread/([0-9]+)/page/([0-9]+)#p([0-9]+)")
var threadURLExpression = regexp.MustCompile("^/forum/view-thread/([0-9]+)")
var forumURLExpression = regexp.MustCompile("^/forum/view-forum/([0-9]+)")

var monthReplacer = strings.NewReplacer(
	"ม.ค.", "Jan",
	"ก.พ.", "Feb",
	"มี.ค.", "Mar",
	"เม.ย.", "Apr",
	"พ.ค.", "May",
	"มิ.ย.", "Jun",
	"ก.ค.", "Jul",
	"ส.ค.", "Aug",
	"ก.ย.", "Sep",
	"ต.ค.", "Oct",
	"พ.ย.", "Nov",
	"ธ.ค.", "Dec",
	"janv.", "Jan",
	"févr.", "Feb",
	"mars", "Mar",
	"avril", "Apr",
	"mai", "May",
	"juin", "Jun",
	"juil.", "Jul",
	"août", "Aug",
	"sept.", "Sep",
	"oct.", "Oct",
	"nov.", "Nov",
	"déc.", "Dec",
)

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

		timeText := monthReplacer.Replace(sel.Find(".post_date").Text())

		for _, format := range []string{
			"Jan _2, 2006 3:04:05 PM",
			"2/1/2006 15:04:05",
			"2.1.2006 15:04:05",
			"_2 Jan 2006, 15:04:05",
			"_2 Jan 2006 15:04:05",
		} {
			if t, err := time.ParseInLocation(format, timeText, timezone); err == nil {
				post.Time = t
				break
			}
		}
		if post.Time.IsZero() {
			log.WithField("text", timeText).Error("unable to parse time")
			return false
		}

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

func (indexer *ForumIndexer) forumPosts(host, poster string, page int, timezone *time.Location) ([]*ForumPost, error) {
	doc, err := indexer.requestDocument(host, fmt.Sprintf("/account/view-posts/%v/page/%v", poster, page))
	if err != nil {
		return nil, err
	}
	posts, err := ScrapeForumPosts(doc, timezone)
	if err != nil {
		return nil, err
	}
	for _, post := range posts {
		post.Host = host
	}
	return posts, nil
}

func (indexer *ForumIndexer) index(host, poster string, timezone *time.Location) {
	logger := log.WithFields(log.Fields{
		"host":   host,
		"poster": poster,
	})

	cutoff := time.Now().Add(time.Hour * -12)
	activity := []Activity(nil)

	for page := 1; ; page++ {
		posts, err := indexer.forumPosts(host, poster, page, timezone)
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

		logger.WithField("count", len(posts)).Info("received forum posts")

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
	doc, err := indexer.requestDocument("www.pathofexile.com", "/my-account/preferences")
	if err != nil {
		return nil, err
	}
	return ScrapeForumTimezone(doc)
}
