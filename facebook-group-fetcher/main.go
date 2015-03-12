package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/huandu/facebook"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"gopkg.in/gorp.v1"
)

func main() {
	// initialize the DbMap
	dbmap := initDb()
	defer dbmap.Db.Close()

	// delete any existing rows
	err := dbmap.TruncateTables()
	checkErr(err, "TruncateTables failed")

	var tokenFile string

	var cmdGroup = &cobra.Command{
		Use:   "group",
		Short: "Fetch group feeds",
		Long:  "Fetch group feeds",
		Run: func(cmd *cobra.Command, args []string) {
			token, err := ioutil.ReadFile(tokenFile)
			checkErr(err, fmt.Sprintf("fail to read token file %v.", tokenFile))
			fetcher := new(GroupFetcher)
			fetcher.SetAccessToken(string(token))
			fetcher.SetDbMap(dbmap)
			fetcher.fetch()
		},
	}

	cmdGroup.Flags().StringVarP(&tokenFile, "token", "t", "", "AccessToken file")

	var rootCmd = &cobra.Command{Use: "facebook-group-fetcher"}
	rootCmd.AddCommand(cmdGroup)
	rootCmd.Execute()
}

// redribbon facebook groupid
const groupId = "258059967652595"

// for fetching feeds
type FBFeed struct {
	Id          string  `facebook:",required"`
	From        *FBUser `facebook:",required"`
	Message     string
	CreatedTime string
	UpdatedTime string
	Comments    []FBComment
	Likes       []FBUser
}

type FBFeeds struct {
	Feeds []FBFeed `facebook:"data,required"`
}

type FBUser struct {
	Id   string `facebook:",required"`
	Name string `facebook:",required"`
}

type FBComment struct {
	Id          string `facebook:",required"`
	FeedId      string
	From        *FBUser `facebook:",required"`
	Message     string  `facebook:",required"`
	LikeCount   string
	CreatedTime string
}

type FBComments struct {
	Comments []FBComment `facebook:"data,required"`
}

type FBLikes struct {
	Users []FBUser `facebook:"data,required"`
}

// for storing on DB
// facebook decoder가 유연하지 않아서 FBFeed의 속성별 데이타형을 수정하기가
// 어려워서 DB용으로 따로 분리함
type Feed struct {
	Id        string `db:"id"`
	From      int64  `db:"from"`
	Message   string `db:"messge"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

type Comment struct {
	Id        int64  `db:"id"`
	FeedId    string `db:"feed_id"`
	From      int64  `db:"from"`
	Message   string `db:"message"`
	LikeCount int64  `db:"like_count"`
	CreatedAt int64  `db:"created_at"`
}

type User struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
}

type Like struct {
	Id     int64  `db:"id"`
	UserId int64  `db:"user_id"`
	FeedId string `db:"feed_id"`
}

type GroupFetcher struct {
	session *facebook.Session
	token   string
	dbmap   *gorp.DbMap
}

func (g *GroupFetcher) SetAccessToken(token string) {
	g.token = token
	g.session = &facebook.Session{}
	g.session.SetAccessToken(g.token)
}

func (g *GroupFetcher) SetDbMap(dbmap *gorp.DbMap) {
	g.dbmap = dbmap
}

func (g *GroupFetcher) fetch() {

	res, err := g.session.Get("/"+groupId+"/feed", facebook.Params{
		"limit":  10,
		"fields": "id,created_time,updated_time,from,message",
	})

	checkErr(err, fmt.Sprintf("fail to get feed. [err:%v]\n", err))

	// create a paging structure.
	paging, _ := res.Paging(g.session)
	noMore := false

	for !noMore {
		var feeds FBFeeds
		err = paging.Decode(&feeds)
		checkErr(err, fmt.Sprintf("fail to decode feeds. [err:%v]\n", err))

		for i, feed := range feeds.Feeds {
			fmt.Printf("Feed %v\n", i)
			fmt.Printf("\tid: %v\n", feed.Id)
			fmt.Printf("\tcreated_time: %v\n", feed.CreatedTime)
			fmt.Printf("\tupdated_time: %v\n", feed.UpdatedTime)
			fmt.Printf("\tmessage: %v\n", feed.Message)
			fmt.Printf("\tfrom: %v\n", feed.From)

			feed.Comments = g.fetchComments(feed.Id)
			feed.Likes = g.fetchLikes(feed.Id)

			g.findOrInsertUser(feed.From)
			g.insertFeed(feed)
		}

		noMore, err = paging.Next()
		checkErr(err, fmt.Sprintf("fail to get feeds' next page. [err:%v]\n", err))
	}
}

func (g *GroupFetcher) fetchComments(feedId string) []FBComment {
	res, err := g.session.Get("/"+feedId+"/comments", facebook.Params{
		"limit": 10,
	})

	checkErr(err, fmt.Sprintf("fail to get comments. [err:%v]\n", err))

	// create a paging structure.
	paging, _ := res.Paging(g.session)

	var result []FBComment

	for noMore := false; !noMore; {
		var comments FBComments
		err = paging.Decode(&comments)
		checkErr(err, fmt.Sprintf("fail to decode comments. [err:%v]\n", err))

		for i, comment := range comments.Comments {
			fmt.Printf("Comment %v\n", i)
			fmt.Printf("\tid: %v\n", comment.Id)
			fmt.Printf("\tcreated_time: %v\n", comment.CreatedTime)
			fmt.Printf("\tmessage: %v\n", comment.Message)
			fmt.Printf("\tfrom: %v\n", comment.From)
			fmt.Printf("\tlike_count: %v\n", comment.LikeCount)

			comment.FeedId = feedId

			g.findOrInsertUser(comment.From)
			g.insertComment(comment)

			result = append(result, comment)
		}

		noMore, err = paging.Next()
		checkErr(err, fmt.Sprintf("fail to get comments' next page. [err:%v]\n", err))
	}

	return result
}

func (g *GroupFetcher) fetchLikes(feedId string) []FBUser {
	res, err := g.session.Get("/"+feedId+"/likes", facebook.Params{
		"limit": 10,
	})

	checkErr(err, fmt.Sprintf("fail to get likes. [err:%v]\n", err))

	// create a paging structure.
	paging, _ := res.Paging(g.session)

	var result []FBUser

	for noMore := false; !noMore; {
		var likes FBLikes
		err = paging.Decode(&likes)
		checkErr(err, fmt.Sprintf("fail to decode likes. [err:%v]\n", err))

		for i, user := range likes.Users {
			fmt.Printf("Like %v\n", i)
			fmt.Printf("\tid: %v\n", user.Id)
			fmt.Printf("\tname: %v\n", user.Name)

			g.findOrInsertUser(&user)

			userId, _ := strconv.ParseInt(user.Id, 10, 64)
			g.insertLike(feedId, userId)

			result = append(result, user)
		}

		noMore, err = paging.Next()
		checkErr(err, fmt.Sprintf("fail to get likes' next page. [err:%v]\n", err))
	}

	return result
}

func (g *GroupFetcher) insertFeed(feed FBFeed) {
	from, _ := strconv.ParseInt(feed.From.Id, 10, 64)
	createdAt, _ := time.Parse("2006-01-02T15:04:05+0000", feed.CreatedTime)
	updatedAt, _ := time.Parse("2006-01-02T15:04:05+0000", feed.UpdatedTime)
	dbFeed := Feed{
		Id:        feed.Id,
		From:      from,
		Message:   feed.Message,
		CreatedAt: createdAt.UnixNano(),
		UpdatedAt: updatedAt.UnixNano(),
	}

	err := g.dbmap.Insert(&dbFeed)
	checkErr(err, "Insert feed failed")
}

func (g *GroupFetcher) insertComment(comment FBComment) {
	id, _ := strconv.ParseInt(comment.Id, 10, 64)
	from, _ := strconv.ParseInt(comment.From.Id, 10, 64)
	likeCount, _ := strconv.ParseInt(comment.LikeCount, 10, 64)
	createdAt, _ := time.Parse("2006-01-02T15:04:05+0000", comment.CreatedTime)
	dbComment := Comment{
		Id:        id,
		FeedId:    comment.FeedId,
		From:      from,
		Message:   comment.Message,
		LikeCount: likeCount,
		CreatedAt: createdAt.UnixNano(),
	}

	err := g.dbmap.Insert(&dbComment)
	checkErr(err, "Insert comment failed")
}

func (g *GroupFetcher) insertLike(feedId string, userId int64) {
	dbLike := Like{
		FeedId: feedId,
		UserId: userId,
	}

	err := g.dbmap.Insert(&dbLike)
	checkErr(err, "Insert like failed")
}

func (g *GroupFetcher) findOrInsertUser(user *FBUser) {
	id, _ := strconv.ParseInt(user.Id, 10, 64)

	obj, err := g.dbmap.Get(User{}, id)
	checkErr(err, "Get user failed")

	// not exists, insert new record
	if obj == nil {
		dbUser := User{
			Id:   id,
			Name: user.Name,
		}

		err = g.dbmap.Insert(&dbUser)
		checkErr(err, "Insert comment failed")
	}
}

func initDb() *gorp.DbMap {
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("sqlite3", "/tmp/facebook_db.bin")
	checkErr(err, "sql.Open failed")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// add a table, setting the table name and PK
	dbmap.AddTableWithName(Feed{}, "feeds").SetKeys(false, "Id")
	dbmap.AddTableWithName(Comment{}, "comments").SetKeys(false, "Id")
	dbmap.AddTableWithName(Like{}, "likes").SetKeys(true, "Id")
	dbmap.AddTableWithName(User{}, "users").SetKeys(false, "Id")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
