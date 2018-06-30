package server

import (
	"net/http"
	"strings"
	"time"
)

type Locale struct {
	Subdomain     string
	Image         string
	IncludeReddit bool
	Translations  map[string]string
	ParseTime     func(s string, tz *time.Location) (time.Time, error)
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
		return a.Host == l.ForumHost()
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

var thMonthReplacer = strings.NewReplacer(
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
)

var frMonthReplacer = strings.NewReplacer(
	"janv.", "Jan",
	"févr.", "Feb",
	"mars", "Mar",
	"avr.", "Apr",
	"mai", "May",
	"juin", "Jun",
	"juil.", "Jul",
	"août", "Aug",
	"sept.", "Sep",
	"oct.", "Oct",
	"nov.", "Nov",
	"déc.", "Dec",
)

// TODO: add translations
var Locales = []*Locale{
	{
		IncludeReddit: true,
		Image:         "static/images/locales/gb.png",
		ParseTime: func(s string, tz *time.Location) (time.Time, error) {
			return time.ParseInLocation("Jan _2, 2006 3:04:05 PM", s, tz)
		},
	},
	{
		Subdomain: "br",
		Image:     "static/images/locales/br.png",
		ParseTime: func(s string, tz *time.Location) (time.Time, error) {
			return time.ParseInLocation("2/1/2006 15:04:05", s, tz)
		},
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
		ParseTime: func(s string, tz *time.Location) (time.Time, error) {
			return time.ParseInLocation("2.1.2006 15:04:05", s, tz)
		},
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
		ParseTime: func(s string, tz *time.Location) (time.Time, error) {
			return time.ParseInLocation("_2 Jan 2006, 15:04:05", thMonthReplacer.Replace(s), tz)
		},
	},
	{
		Subdomain: "de",
		Image:     "static/images/locales/de.png",
		ParseTime: func(s string, tz *time.Location) (time.Time, error) {
			return time.ParseInLocation("2.1.2006 15:04:05", s, tz)
		},
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
		ParseTime: func(s string, tz *time.Location) (time.Time, error) {
			return time.ParseInLocation("_2 Jan 2006 15:04:05", frMonthReplacer.Replace(s), tz)
		},
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
		ParseTime: func(s string, tz *time.Location) (time.Time, error) {
			return time.ParseInLocation("2/1/2006 15:04:05", s, tz)
		},
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
