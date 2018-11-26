package main

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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
	pflag.String("db", "./gggtracker.db", "the database file path")
	pflag.String("dynamodb-table", "", "if given, DynamoDB will be used instead of a database file")
	pflag.String("forumsession", "", "the POESESSID cookie for a forum session")
	viper.BindPFlags(pflag.CommandLine)
	pflag.Parse()

	viper.SetEnvPrefix("gggtracker")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	e := echo.New()

	var db server.Database
	var err error
	if tableName := viper.GetString("dynamodb-table"); tableName != "" {
		config, err := external.LoadDefaultAWSConfig()
		if err != nil {
			log.Fatal(err)
		}
		db, err = server.NewDynamoDBDatabase(dynamodb.New(config), tableName)
	} else {
		db, err = server.NewBoltDatabase(viper.GetString("db"))
	}
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
