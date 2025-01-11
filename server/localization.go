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
	HelpForumId   int

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
		HelpForumId:   584,
	},
	{
		Subdomain: "br",
		Image:     "static/images/locales/br.png",
		Translations: map[string]string{
			"Activity":        "Atividade",
			"Thread":          "Discussão",
			"Poster":          "Autor",
			"Time":            "Hora",
			"Forum":           "Fórum",
			"Hide Help Forum": "Ocultar Fórum de Ajuda",
			"Show Help Forum": "Mostrar Fórum de Ajuda",
		},
		HelpForumId: 774,
	},
	{
		Subdomain: "ru",
		Image:     "static/images/locales/ru.png",
		Translations: map[string]string{
			"Activity":        "Активность",
			"Thread":          "Тема",
			"Poster":          "Автор",
			"Time":            "Время",
			"Forum":           "Форум",
			"Hide Help Forum": "Скрыть форум помощи",
			"Show Help Forum": "Показать Форум Помощи",
		},
		HelpForumId: 1281,
	},
	{
		Subdomain: "th",
		Image:     "static/images/locales/th.png",
		Translations: map[string]string{
			"Activity":        "กิจกรรม",
			"Thread":          "กระทู้",
			"Poster":          "ผู้โพสต์",
			"Time":            "เวลา",
			"Forum":           "ฟอรั่ม",
			"Hide Help Forum": "ซ่อนฟอรั่มช่วยเหลือ",
			"Show Help Forum": "แสดงฟอรั่มช่วยเหลือ",
		},
		HelpForumId: 1011,
	},
	{
		Subdomain: "de",
		Image:     "static/images/locales/de.png",
		Translations: map[string]string{
			"Activity":        "Aktivität",
			"Thread":          "Beitrag",
			"Poster":          "Verfasser",
			"Time":            "Datum",
			"Forum":           "Forum",
			"Hide Help Forum": "Hilfeforum ausblenden",
			"Show Help Forum": "Hilfeforum anzeigen",
		},
		HelpForumId: 1123,
	},
	{
		Subdomain: "fr",
		Image:     "static/images/locales/fr.png",
		Translations: map[string]string{
			"Activity":        "Activité",
			"Thread":          "Fil de discussion",
			"Poster":          "Posteur",
			"Time":            "Date",
			"Forum":           "Forum",
			"Hide Help Forum": "Masquer le forum d'aide",
			"Show Help Forum": "Afficher le forum d'aide",
		},
		HelpForumId: 1051,
	},
	{
		Subdomain: "es",
		Image:     "static/images/locales/es.png",
		Translations: map[string]string{
			"Activity":        "Actividad",
			"Thread":          "Tema",
			"Poster":          "Autor",
			"Time":            "Fecha",
			"Forum":           "Foro",
			"Hide Help Forum": "Ocultar el foro de ayuda",
			"Show Help Forum": "Mostrar el foro de ayuda",
		},
		HelpForumId: 1193,
	},
	{
		Subdomain: "jp",
		Image:     "static/images/locales/jp.png",
		Translations: map[string]string{
			"Activity": "アクティビティ",
			"Thread":   "スレッド",
			"Poster":   "投稿者",
			"Time":     "日時",
			"Forum":    "フォーラム",
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
