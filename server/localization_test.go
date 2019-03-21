package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocale_RefreshForumIds(t *testing.T) {
	for _, l := range Locales {
		t.Run(l.ForumHost(), func(t *testing.T) {
			assert.NoError(t, l.RefreshForumIds())
			assert.NotEmpty(t, l.ForumIds())
		})
	}
}
