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
	log "github.com/sirupsen/logrus"
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

type Server struct {
	*echo.Echo

	close  chan struct{}
	closed chan struct{}
}

func New(db Database, ga string) *Server {
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

	ret := &Server{
		Echo:   e,
		close:  make(chan struct{}),
		closed: make(chan struct{}),
	}
	go ret.run()
	return ret
}

func (s *Server) Close() error {
	close(s.close)
	<-s.closed
	return s.Echo.Close()
}

func (s *Server) run() {
	defer close(s.closed)

	for {
		for _, locale := range Locales {
			select {
			case <-s.close:
				return
			default:
				logger := log.WithField("host", locale.ForumHost())
				if err := locale.RefreshForumIds(); err != nil {
					logger.WithError(err).Error("error refreshing forum ids")
				} else {
					logger.Info("refreshed forum ids")
				}
				time.Sleep(time.Second)
			}
		}
		select {
		case <-s.close:
			return
		case <-time.After(time.Minute):
		}
	}
}
