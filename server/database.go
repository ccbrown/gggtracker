package server

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io/ioutil"

	json "github.com/json-iterator/go"
)

const (
	ForumPostType = iota
	RedditCommentType
	RedditPostType
)

type Database interface {
	AddActivity(activity []Activity) error
	Activity(locale *Locale, start string, count int, filter func(a Activity) bool) ([]Activity, string, error)
	Close() error
}

const gzipMarker = 0

func marshalActivity(a Activity) (key, value []byte, err error) {
	buf := &bytes.Buffer{}
	buf.Write([]byte{gzipMarker})
	w := gzip.NewWriter(buf)
	if err := json.NewEncoder(w).Encode(a); err != nil {
		return nil, nil, err
	}
	w.Close()
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
	return k, buf.Bytes(), nil
}

func unmarshalActivity(key, value []byte) (Activity, error) {
	if len(value) > 0 && value[0] == gzipMarker {
		r, err := gzip.NewReader(bytes.NewReader(value[1:]))
		if err != nil {
			return nil, err
		}
		defer r.Close()
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		value = buf
	}

	switch key[5] {
	case ForumPostType:
		post := &ForumPost{}
		err := json.Unmarshal(value, post)
		if err != nil {
			return nil, err
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
