package server

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"

	"github.com/boltdb/bolt"
)

type Database struct {
	db *bolt.DB
}

func OpenDatabase(path string) (*Database, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("activity"))
		if err != nil {
			return err
		}
		return nil
	})

	return &Database{
		db: db,
	}, nil
}

func (db *Database) Close() {
	db.db.Close()
}

const (
	ForumPostType = iota
	RedditCommentType
	RedditPostType
)

func (db *Database) AddActivity(activity []Activity) {
	err := db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("activity"))
		for _, a := range activity {
			buf, err := json.Marshal(a)
			if err != nil {
				return err
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
			b.Put(k, buf)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (db *Database) Activity(start string, count int, filter func(Activity) bool) ([]Activity, string) {
	ret := []Activity(nil)
	next := ""
	err := db.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("activity")).Cursor()
		var k, v []byte
		if start == "" {
			k, v = c.Last()
		} else {
			s, err := base64.RawURLEncoding.DecodeString(start)
			if err != nil {
				k, v = c.Last()
			} else {
				k, v = c.Seek(s)
				if k != nil {
					k, v = c.Prev()
				}
			}
		}
		for len(ret) < count && k != nil {
			var activity Activity
			switch k[5] {
			case ForumPostType:
				post := &ForumPost{}
				err := json.Unmarshal(v, post)
				if err != nil {
					return err
				}
				if post.Host == "" {
					post.Host = "www.pathofexile.com"
				}
				activity = post
			case RedditCommentType:
				comment := &RedditComment{}
				err := json.Unmarshal(v, comment)
				if err != nil {
					return err
				}
				activity = comment
			case RedditPostType:
				post := &RedditPost{}
				err := json.Unmarshal(v, post)
				if err != nil {
					return err
				}
				activity = post
			}
			if filter == nil || filter(activity) {
				ret = append(ret, activity)
				next = base64.RawURLEncoding.EncodeToString(k)
			}
			k, v = c.Prev()
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return ret, next
}
