package server

import (
	"encoding/base64"

	"github.com/boltdb/bolt"
)

type BoltDatabase struct {
	db *bolt.DB
}

func NewBoltDatabase(path string) (*BoltDatabase, error) {
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

	return &BoltDatabase{
		db: db,
	}, nil
}

func (db *BoltDatabase) AddActivity(activity []Activity) error {
	return db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("activity"))
		for _, a := range activity {
			k, v, err := marshalActivity(a)
			if err != nil {
				return err
			}
			b.Put(k, v)
		}
		return nil
	})
}

func (db *BoltDatabase) Activity(locale *Locale, start string, count int, filter func(a Activity) bool) ([]Activity, string, error) {
	ret := []Activity(nil)
	next := ""
	if err := db.db.View(func(tx *bolt.Tx) error {
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
			activity, err := unmarshalActivity(k, v)
			if err != nil {
				return err
			} else if activity != nil && locale.ActivityFilter(activity) && (filter == nil || filter(activity)) {
				ret = append(ret, activity)
				next = base64.RawURLEncoding.EncodeToString(k)
			}
			k, v = c.Prev()
		}
		return nil
	}); err != nil {
		return nil, "", err
	}
	return ret, next, nil
}

func (db *BoltDatabase) Close() error {
	return db.db.Close()
}
