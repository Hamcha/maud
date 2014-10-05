package main

import (
	"gopkg.in/mgo.v2/bson"
)

type SiteInfo struct {
	Title  string
	Footer string
}

type User struct {
	Nickname string
	Tripcode string
}

type Thread struct {
	Id        bson.ObjectId "_id"
	ShortUrl  string
	Title     string
	Author    User
	Content   string
	Tags      []string
	Date      int64
	Messages  int32
	LastReply bson.ObjectId
	LRDate    int64
}

type Post struct {
	Id       bson.ObjectId "_id"
	ThreadId bson.ObjectId
	Author   User
	Content  string
	Date     int64
}
