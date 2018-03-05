package server

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_ForumPosts(t *testing.T) {
	dir, err := ioutil.TempDir("testdata", "db")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := OpenDatabase(path.Join(dir, "test.db"))
	require.NoError(t, err)
	defer db.Close()

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

	posts, next := db.Activity("", 1, nil)
	require.Equal(t, 1, len(posts))
	assert.Equal(t, post1.Id, posts[0].(*ForumPost).Id)
	assert.Equal(t, post1.Poster, posts[0].(*ForumPost).Poster)
	assert.Equal(t, post1.Time.Unix(), posts[0].(*ForumPost).Time.Unix())

	posts, next = db.Activity(next, 1, nil)
	require.Equal(t, 1, len(posts))
	assert.Equal(t, post2.Id, posts[0].(*ForumPost).Id)
	assert.Equal(t, post2.Poster, posts[0].(*ForumPost).Poster)
	assert.Equal(t, post2.Time.Unix(), posts[0].(*ForumPost).Time.Unix())

	posts, _ = db.Activity(next, 1, nil)
	require.Equal(t, 0, len(posts))
}
