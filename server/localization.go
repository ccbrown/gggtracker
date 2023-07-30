package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Locale struct {
	Subdomain     string
	Image         string
	IncludeReddit bool
	Translations  map[string]string
	ParseTime     func(s string, tz *time.Location) (time.Time, error)

	forumIds atomic.Value
}

func (l *Locale) Translate(s string) string {
	if translated, ok := l.Translations[s]; ok {
		return translated
	}
	return s
}

func (l *Locale) ActivityFilter(a Activity) bool {
	switch a := a.(type) {
	case *ForumPost:
		return a.Host() == l.ForumHost()
	case *RedditComment:
		return l.IncludeReddit
	case *RedditPost:
		return l.IncludeReddit
	}
	return false
}

func (l *Locale) ForumHost() string {
	if l.Subdomain != "" {
		return l.Subdomain + ".pathofexile.com"
	}
	return "www.pathofexile.com"
}

func (l *Locale) ForumIds() map[int]bool {
	ret, _ := l.forumIds.Load().(map[int]bool)
	return ret
}

func (l *Locale) RefreshForumIds() error {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%v/forum", l.ForumHost()), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	forumIds := map[int]bool{}
	doc.Find(".forumTable tbody tr").Each(func(i int, sel *goquery.Selection) {
		if idStr := sel.AttrOr("data-id", ""); idStr != "" {
			if id, err := strconv.Atoi(idStr); err == nil {
				forumIds[id] = true
			}
		}
	})
	l.forumIds.Store(forumIds)

	return nil
}

var Locales = []*Locale{
	{
		IncludeReddit: true,
		Image:         "static/images/locales/gb.png",
	},
	{
		Subdomain: "br",
		Image:     "static/images/locales/br.png",
		Translations: map[string]string{
			"Activity": "Atividade",
			"Thread":   "Discussão",
			"Poster":   "Autor",
			"Time":     "Hora",
			"Forum":    "Fórum",
		},
	},
	{
		Subdomain: "ru",
		Image:     "static/images/locales/ru.png",
		Translations: map[string]string{
			"Activity": "Активность",
			"Thread":   "Тема",
			"Poster":   "Автор",
			"Time":     "Время",
			"Forum":    "Форум",
		},
	},
	{
		Subdomain: "th",
		Image:     "static/images/locales/th.png",
	},
	{
		Subdomain: "de",
		Image:     "static/images/locales/de.png",
		Translations: map[string]string{
			"Activity": "Aktivität",
			"Thread":   "Beitrag",
			"Poster":   "Verfasser",
			"Time":     "Datum",
			"Forum":    "Forum",
		},
	},
	{
		Subdomain: "fr",
		Image:     "static/images/locales/fr.png",
		Translations: map[string]string{
			"Activity": "Activité",
			"Thread":   "Fil de discussion",
			"Poster":   "Posteur",
			"Time":     "Date",
			"Forum":    "Forum",
		},
	},
	{
		Subdomain: "es",
		Image:     "static/images/locales/es.png",
		Translations: map[string]string{
			"Activity": "Actividad",
			"Thread":   "Tema",
			"Poster":   "Autor",
			"Time":     "Fecha",
			"Forum":    "Foro",
		},
	},
}

func LocaleForRequest(r *http.Request) *Locale {
	subdomain := ""
	if r.Host != "" {
		subdomain = strings.Split(r.Host, ".")[0]
	}

	if subdomain != "" {
		for _, l := range Locales {
			if l.Subdomain == subdomain {
				return l
			}
		}
	}

	return Locales[0]
}
