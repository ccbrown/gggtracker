package server

import (
	"os"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScrapeForumPosts(t *testing.T) {
	f, err := os.Open("testdata/forum-posts.html")
	require.NoError(t, err)
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	tz, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)

	poster := ForumAccount{Username: "Chris"}
	posts, err := ScrapeForumPosts(doc, poster, tz)
	require.NoError(t, err)
	assert.Equal(t, 10, len(posts))

	require.NotEmpty(t, posts)
	p := posts[0]
	assert.Equal(t, 14168107, p.Id)
	assert.Equal(t, 54, p.ForumId)
	assert.Equal(t, 1830139, p.ThreadId)
	assert.Equal(t, "Chris", p.Poster)
	assert.Equal(t, 0, p.PosterDiscriminator)
	assert.Equal(t, "Photos of the Fan Meetup", p.ThreadTitle)
	assert.Equal(t, "Announcements", p.ForumName)
	assert.Equal(t, "we had a great ti<strong>m</strong>e too!", p.BodyHTML)
	assert.Equal(t, int64(1486332365), p.Time.Unix())

	t.Run("LoggedOut", func(t *testing.T) {
		f, err := os.Open("testdata/forum-posts-logged-out.html")
		require.NoError(t, err)
		defer f.Close()

		doc, err := goquery.NewDocumentFromReader(f)
		require.NoError(t, err)

		tz, err := time.LoadLocation("America/Los_Angeles")
		require.NoError(t, err)

		posts, err := ScrapeForumPosts(doc, poster, tz)
		require.Error(t, err)
		assert.Equal(t, 0, len(posts))
	})

	t.Run("Maintenance", func(t *testing.T) {
		f, err := os.Open("testdata/forum-maintenance.html")
		require.NoError(t, err)
		defer f.Close()

		doc, err := goquery.NewDocumentFromReader(f)
		require.NoError(t, err)

		tz, err := time.LoadLocation("America/Los_Angeles")
		require.NoError(t, err)

		posts, err := ScrapeForumPosts(doc, poster, tz)
		assert.Equal(t, err, ErrForumMaintenance)
		assert.Equal(t, 0, len(posts))
	})
}

func TestScrapeForumTimezone(t *testing.T) {
	f, err := os.Open("testdata/forum-preferences.html")
	require.NoError(t, err)
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	tz, err := ScrapeForumTimezone(doc)
	require.NoError(t, err)
	assert.Equal(t, "America/Los_Angeles", tz.String())
}
