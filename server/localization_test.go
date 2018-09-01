package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLocale_ParseTime(t *testing.T) {
	for subdomain, s := range map[string]string{
		"":   "Aug 29, 2018, 5:51:19 PM",
		"br": "31 de ago de 2018 00:50:19",
		"ru": "1 сент. 2018 г., 2:09:52",
		"th": "31 ส.ค. 2018 00:50:25",
		"de": "31.08.2018, 00:50:20",
		"fr": "31 août 2018 00:50:22",
		"es": "31 ago. 2018 0:50:23",
	} {
		t.Run(subdomain, func(t *testing.T) {
			var locale *Locale
			for _, l := range Locales {
				if l.Subdomain == subdomain {
					locale = l
					break
				}
			}
			_, err := locale.ParseTime(s, time.FixedZone("UTC-5", -5*60*60))
			assert.NoError(t, err)
		})
	}
}
