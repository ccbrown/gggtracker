package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDatabase(t *testing.T, db Database) {
	t.Run("ForumPosts", func(t *testing.T) {
		testDatabase_ForumPosts(t, db)
	})
}

func testDatabase_ForumPosts(t *testing.T, db Database) {
	locale := Locales[0]

	post1 := &ForumPost{
		Id:     9000,
		Poster: "Chris",
		Time:   time.Unix(1486332365, 0),
	}

	post2 := &ForumPost{
		Id:     9001,
		Poster: "Chris",
		Time:   time.Unix(1486332364, 0),
	}

	db.AddActivity([]Activity{post1, post2})

	all := func(a Activity) bool { return true }

	posts, next, err := db.Activity(locale, "", 1, all)
	require.NoError(t, err)
	require.Equal(t, 1, len(posts))
	assert.Equal(t, post1.Id, posts[0].(*ForumPost).Id)
	assert.Equal(t, post1.Poster, posts[0].(*ForumPost).Poster)
	assert.Equal(t, post1.Time.Unix(), posts[0].(*ForumPost).Time.Unix())

	posts, next, err = db.Activity(locale, next, 1, all)
	require.NoError(t, err)
	require.Equal(t, 1, len(posts))
	assert.Equal(t, post2.Id, posts[0].(*ForumPost).Id)
	assert.Equal(t, post2.Poster, posts[0].(*ForumPost).Poster)
	assert.Equal(t, post2.Time.Unix(), posts[0].(*ForumPost).Time.Unix())

	posts, _, err = db.Activity(locale, next, 1, all)
	require.NoError(t, err)
	require.Equal(t, 0, len(posts))
}
