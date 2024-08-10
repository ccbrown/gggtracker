package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/ccbrown/gggtracker/server"
)

func main() {
	pflag.IntP("port", "p", 8080, "the port to listen on")
	pflag.String("staticdir", "", "this argument is ignored and will be removed")
	pflag.String("ga", "", "a google analytics account")
	pflag.String("db", "./gggtracker.db", "the database file path")
	pflag.String("dynamodb-table", "", "if given, DynamoDB will be used instead of a database file")
	pflag.String("forumsession", "", "the POESESSID cookie for a forum session")
	pflag.String("reddit-auth", "", "the APPLICATION:SECRET to use as Reddit auth")
	viper.BindPFlags(pflag.CommandLine)
	pflag.Parse()

	viper.SetEnvPrefix("gggtracker")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

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

	if viper.GetString("reddit-auth") != "" {
		redditIndexer, err := server.NewRedditIndexer(server.RedditIndexerConfiguration{
			Database: db,
			Auth:     viper.GetString("reddit-auth"),
		})
		if err != nil {
			log.Fatal(err)
		}
		defer redditIndexer.Close()
	}

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

	e := server.New(db, viper.GetString("ga"))
	log.Fatal(e.Start(fmt.Sprintf(":%v", viper.GetInt("port"))))
}
