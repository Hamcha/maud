package main

import (
	"gopkg.in/mgo.v2/bson"
)

type SiteInfo struct {
	Title  string
	Secret string
	Footer string
}

type User struct {
	Nickname string
	Tripcode string
}

type Thread struct {
	Id         bson.ObjectId "_id"
	ShortUrl   string
	Title      string
	Author     User
	Tags       []string
	Date       int64
	Messages   int32
	ThreadPost bson.ObjectId
	LastReply  bson.ObjectId
	LRDate     int64
}

type Post struct {
	Id           bson.ObjectId "_id"
	ThreadId     bson.ObjectId
	Author       User
	Content      string
	Date         int64
	LastModified int64
	ContentType  string
}

type Counter struct {
	Name string
	Seq  int64
}

type Tag struct {
	Name       string
	Posts      int64
	LastUpdate int64
	LastThread bson.ObjectId
}

type TagData struct {
	Name       string
	LastUpdate int64
	LastThread Thread
	LastIndex  int64
}
