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
	tid := bson.NewObjectId()
	pid := bson.NewObjectId()
	now := time.Now().UTC().Unix()

	thread := Thread{
		Id:         tid,
		ShortUrl:   generateURL("thread"),
		Title:      title,
		Author:     user,
		Tags:       tags,
		Date:       now,
		Messages:   1,
		ThreadPost: pid,
		LastReply:  pid,
		LRDate:     now,
	}

	post := Post{
		Id:       pid,
		ThreadId: tid,
		Author:   user,
		Content:  content,
		Date:     now,
	}

	err := database.C("threads").Insert(thread)
	if err != nil {
		return "", err
	}
	err = database.C("posts").Insert(post)
	return thread.ShortUrl, err
}

func DBGetThread(surl string) (Thread, error) {
	var thread Thread
	err := database.C("threads").Find(bson.M{"ShortUrl": surl}).One(&thread)
	return thread, err
}

func DBGetPosts(thread *Thread, limit int, offset int) ([]Post, error) {
	query := database.C("posts").Find(bson.M{"ThreadId": thread.Id})
	if offset > 0 {
		query = query.Skip(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	var posts []Post
	err := query.All(&posts)
	return posts, err
}

func DBNextId(name string) (int64, error) {
	inc := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"Seq": 1}},
		Upsert:    true,
		ReturnNew: true,
	}
	var doc Counter
	_, err := database.C("counters").Find(bson.M{"Name": name}).Apply(inc, &doc)
	return doc.Seq, err
}
