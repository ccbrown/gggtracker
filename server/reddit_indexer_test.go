package server

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRedditComments(t *testing.T) {
	f, err := os.Open("testdata/reddit-user.json")
	require.NoError(t, err)
	defer f.Close()

	json, err := ioutil.ReadAll(f)

	activity, next, err := ParseRedditActivity(json)
	require.NoError(t, err)
	assert.Equal(t, "t1_dcjasef", next)

	require.Equal(t, 24, len(activity))
	post, ok := activity[0].(*RedditPost)
	require.True(t, ok)
	assert.Equal(t, "5q12qc", post.Id)
	assert.Equal(t, "chris_wilson", post.Author)
	assert.Equal(t, "https://www.reddit.com/r/pathofexile/comments/5q12qc/another_update_on_singapore_fibre_cuts/", post.URL)
	assert.Equal(t, "/r/pathofexile/comments/5q12qc/another_update_on_singapore_fibre_cuts/", post.Permalink)
	assert.Equal(t, "Another Update on Singapore Fibre Cuts", post.Title)
	assert.Equal(t, time.Unix(1485316926, 0), post.Time)
	assert.Equal(t, "pathofexile", post.Subreddit)

	comment, ok := activity[1].(*RedditComment)
	require.True(t, ok)
	assert.Equal(t, "chris_wilson", comment.Author)
	assert.Equal(t, "5plxw0", comment.PostId)
	assert.Equal(t, "Development Manifesto: Solo Self-Found Support in 2.6.0", comment.PostTitle)
	assert.Equal(t, time.Unix(1485287813, 0), comment.Time)
	assert.Equal(t, "pathofexile", comment.Subreddit)
}
