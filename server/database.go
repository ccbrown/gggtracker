package server

import (
	"encoding/binary"

	json "github.com/json-iterator/go"
)

const (
	ForumPostType = iota
	RedditCommentType
	RedditPostType
)

type Database interface {
	AddActivity(activity []Activity) error
	Activity(locale *Locale, start string, count int) ([]Activity, string, error)
	Close() error
}

func marshalActivity(a Activity) (key, value []byte, err error) {
	buf, err := json.Marshal(a)
	if err != nil {
		return nil, nil, err
	}
	k := make([]byte, 10)
	binary.BigEndian.PutUint64(k, uint64(a.ActivityTime().Unix())<<24)
	switch a.(type) {
	case *ForumPost:
		k[5] = ForumPostType
	case *RedditComment:
		k[5] = RedditCommentType
	case *RedditPost:
		k[5] = RedditPostType
	}
	binary.BigEndian.PutUint32(k[6:], a.ActivityKey())
	return k, buf, nil
}

func unmarshalActivity(key, value []byte) (Activity, error) {
	switch key[5] {
	case ForumPostType:
		post := &ForumPost{}
		err := json.Unmarshal(value, post)
		if err != nil {
			return nil, err
		}
		if post.Host == "" {
			post.Host = "www.pathofexile.com"
		}
		if post.Id != 0 {
			return post, nil
		}
	case RedditCommentType:
		comment := &RedditComment{}
		err := json.Unmarshal(value, comment)
		if err != nil {
			return nil, err
		}
		return comment, nil
	case RedditPostType:
		post := &RedditPost{}
		err := json.Unmarshal(value, post)
		if err != nil {
			return nil, err
		}
		return post, nil
	}
	return nil, nil
}
