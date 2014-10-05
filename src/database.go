package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var session *mgo.Session
var database *mgo.Database

func DBInit(servers, dbname string) {
	var err error
	session, err = mgo.Dial(servers)
	if err != nil {
		panic(err)
	}
	database = session.DB(dbname)
}

func DBClose() {
	session.Close()
}

func DBNewThread(user User, title, content string, tags []string) (string, error) {
	now := time.Now().UTC().Unix()
	thread := Thread{
		Id:       bson.NewObjectId(),
		ShortUrl: generateURL(now),
		Title:    title,
		Content:  content,
		Tags:     tags,
		Author:   user,
		Date:     now,
		Messages: 1,
	}
	err := database.C("threads").Insert(thread)
	return thread.ShortUrl, err
}
