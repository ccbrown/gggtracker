package main

import (
	"fmt"
	"net/http"
	"path"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/ccbrown/gggtracker/server"
)

func main() {
	pflag.IntP("port", "p", 8080, "the port to listen on")
	pflag.String("staticdir", "./server/static", "the static files to serve")
	pflag.String("ga", "", "a google analytics account")
	pflag.String("db", "./gggtracker.db", "the database path")
	pflag.String("forumsession", "", "the POESESSID cookie for a forum session")
	viper.BindPFlags(pflag.CommandLine)
	pflag.Parse()

	viper.SetEnvPrefix("gggtracker")
	viper.AutomaticEnv()

	e := echo.New()

	db, err := server.OpenDatabase(viper.GetString("db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	redditIndexer, err := server.NewRedditIndexer(server.RedditIndexerConfiguration{
		Database: db,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer redditIndexer.Close()

	if viper.GetString("forumsession") != "" {
		forumIndexer, err := server.NewForumIndexer(server.ForumIndexerConfiguration{
			Database: db,
			Session:  viper.GetString("forumsession"),
		})
		if err != nil {
			log.Fatal(err)
		}
		defer forumIndexer.Close()
	}

	e.Use(middleware.Recover())

	e.GET("/", server.IndexHandler(server.IndexConfiguration{
		GoogleAnalytics: viper.GetString("ga"),
	}))
	e.GET("/activity.json", server.ActivityHandler(db))
	e.GET("/rss", server.RSSHandler(db))
	e.GET("/rss.php", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, server.AbsoluteURL(c, "/rss"))
	})
	e.File("/favicon.ico", path.Join(viper.GetString("staticdir"), "favicon.ico"))
	e.Static("/static", viper.GetString("staticdir"))

	log.Fatal(e.Start(fmt.Sprintf(":%v", viper.GetInt("port"))))
}
