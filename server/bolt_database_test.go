package server

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBoltDatabase(t *testing.T) {
	dir, err := ioutil.TempDir("testdata", "db")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := NewBoltDatabase(path.Join(dir, "test.db"))
	require.NoError(t, err)
	defer db.Close()

	testDatabase(t, db)
}
