package main

import (
	"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
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
