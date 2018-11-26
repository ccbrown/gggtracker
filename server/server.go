//go:generate sh -c "go get -u github.com/kevinburke/go-bindata/... && `go env GOPATH`/bin/go-bindata -pkg server -ignore '(^|/)\\..*' static/..."
package server

import (
	"bytes"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func serveAsset(c echo.Context, path string) error {
	b, err := Asset(path)
	if err != nil {
		http.NotFound(c.Response(), c.Request())
		return nil
	}
	http.ServeContent(c.Response(), c.Request(), path, time.Time{}, bytes.NewReader(b))
	return nil
}

func New(db Database, ga string) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Recover())

	e.GET("/", IndexHandler(IndexConfiguration{
		GoogleAnalytics: ga,
	}))
	e.GET("/activity.json", ActivityHandler(db))
	e.GET("/rss", RSSHandler(db))
	e.GET("/rss.php", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, AbsoluteURL(c, "/rss"))
	})
	e.GET("/favicon.ico", func(c echo.Context) error {
		return serveAsset(c, "static/favicon.ico")
	})
	e.GET("/static/*", func(c echo.Context) error {
		p, err := url.PathUnescape(c.Param("*"))
		if err != nil {
			return err
		}
		return serveAsset(c, filepath.Join("static", path.Clean("/"+p)))
	})

	return e
}
