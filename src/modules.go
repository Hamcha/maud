package main

import (
	// Database
	"./database/mongo"
	// Formatters
	"./modules"
	"./modules/formatters/bbcode"
	"./modules/formatters/markdown"
)

var formatters []modules.Formatter

func InitFormatters() {
	formatters = make([]modules.Formatter, 0)
	formatters = append(formatters, bbcode.Provide())
	formatters = append(formatters, markdown.Provide())
}

var database Database

type User mongo.User
type Thread mongo.Thread
type ThreadInfo mongo.ThreadInfo
type Post mongo.Post
type Tag mongo.Tag
type TagData mongo.TagData

func InitDatabase(servers, dbname string) {
	database := mongo.Init(servers, dbname)
}
