package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type RedditIndexerConfiguration struct {
	Database Database
	Auth     string
}

type RedditIndexer struct {
	configuration RedditIndexerConfiguration
	closeSignal   chan bool
}

func NewRedditIndexer(configuration RedditIndexerConfiguration) (*RedditIndexer, error) {
	ret := &RedditIndexer{
		configuration: configuration,
		closeSignal:   make(chan bool),
	}
	go ret.run()
	return ret, nil
}

func (indexer *RedditIndexer) Close() {
	indexer.closeSignal <- true
}

const redditRequestInterval = time.Second * 8

func (indexer *RedditIndexer) run() {
	log.Info("starting reddit indexer")

	users := []string{
		"chris_wilson", "Bex_GGG", "Negitivefrags", "Omnitect", "Mark_GGG", "RhysGGG", "Dan_GGG",
		"Rory_Rackham", "Blake_GGG", "Fitzy_GGG", "Hartlin_GGG", "Hrishi_GGG", "Baltic_GGG",
		"KamilOrmanJanowski", "Daniel_GGG", "Jeff_GGG", "NapfelGGG", "Baltic_GGG", "Novynn",
		"Felipe_GGG", "Mel_GGG", "Sarah_GGG", "riandrake", "Kieren_GGG", "Openarl", "Natalia_GGG",
		"AlexDenfordGGG", "Stacey_GGG", "ZaccieA", "viperesque", "rach_ggg", "Community_Team",
		"M59Gar", "Dominic_GGG", "Nick_GGG", "MatthewD_GGG",
	}
	next := 0

	for {
		select {
		case <-indexer.closeSignal:
			return
		default:
			if err := indexer.index(users[next]); err != nil {
				log.WithError(err).Error("error indexing reddit user: " + users[next])
			}
			next += 1
			if next >= len(users) {
				next = 0
			}
			time.Sleep(redditRequestInterval)
		}
	}
}

func ParseRedditActivity(b []byte) ([]Activity, string, error) {
	activity := []Activity(nil)

	var root struct {
		Data struct {
			After    string `json:"after"`
			Children []struct {
				Kind string `json:"kind"`
				Data struct {
					Id           string  `json:"id"`
					Author       string  `json:"author"`
					BodyHTML     string  `json:"body_html"`
					SelftextHTML string  `json:"selftext_html"`
					SubredditId  string  `json:"subreddit_id"`
					Permalink    string  `json:"permalink"`
					URL          string  `json:"url"`
					Title        string  `json:"title"`
					CreatedUTC   float64 `json:"created_utc"`
					LinkId       string  `json:"link_id"`
					LinkTitle    string  `json:"link_title"`
					Subreddit    string  `json:"subreddit"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	if err := json.Unmarshal(b, &root); err != nil {
		return nil, "", err
	}

	for _, thing := range root.Data.Children {
		if thing.Data.SubredditId != "t5_2sf6m" && thing.Data.SubredditId != "t5_2w3q8" && thing.Data.SubredditId != "t5_3910n" {
			continue
		}
		switch thing.Kind {
		case "t1":
			activity = append(activity, &RedditComment{
				Id:        thing.Data.Id,
				Author:    thing.Data.Author,
				BodyHTML:  thing.Data.BodyHTML,
				PostId:    strings.TrimPrefix(thing.Data.LinkId, "t3_"),
				PostTitle: thing.Data.LinkTitle,
				Time:      time.Unix(int64(thing.Data.CreatedUTC), 0),
				Subreddit: thing.Data.Subreddit,
			})
		case "t3":
			activity = append(activity, &RedditPost{
				Id:        thing.Data.Id,
				Author:    thing.Data.Author,
				BodyHTML:  thing.Data.SelftextHTML,
				Permalink: thing.Data.Permalink,
				Title:     thing.Data.Title,
				URL:       thing.Data.URL,
				Time:      time.Unix(int64(thing.Data.CreatedUTC), 0),
				Subreddit: thing.Data.Subreddit,
			})
		}
	}

	return activity, root.Data.After, nil
}

func (indexer *RedditIndexer) redditActivity(user string, page string) ([]Activity, string, error) {
	url := fmt.Sprintf("https://www.reddit.com/user/%v.json?count=25&after=%v&raw_json=1", user, page)
	client := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Add("User-Agent", "GGG Tracker (https://github.com/ccbrown/gggtracker) by /u/rz2yoj")
	if parts := strings.Split(indexer.configuration.Auth, ":"); len(parts) == 2 {
		req.SetBasicAuth(parts[0], parts[1])
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return ParseRedditActivity(body)
}

func (indexer *RedditIndexer) index(user string) error {
	logger := log.WithFields(log.Fields{
		"user": user,
	})

	pageCutoff := time.Now().Add(-12 * time.Hour)
	cutoff := time.Now().Add(-14 * 24 * time.Hour)
	activity := []Activity(nil)

	for page := ""; ; {
		things, next, err := indexer.redditActivity(user, page)
		page = next
		if err != nil {
			logger.WithError(err).Error("error requesting reddit activity")
		}

		done := len(things) == 0 || next == ""
		for _, thing := range things {
			if thing.ActivityTime().Before(pageCutoff) {
				done = true
			}
			if thing.ActivityTime().Before(cutoff) {
				break
			}
			activity = append(activity, thing)
		}

		logger.WithFields(log.Fields{
			"count": len(activity),
		}).Info("received reddit activity")

		if done {
			break
		}
		time.Sleep(redditRequestInterval)
	}

	if len(activity) == 0 {
		return nil
	}

	return indexer.configuration.Database.AddActivity(activity)
}
