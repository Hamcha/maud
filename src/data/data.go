package data

import (
	"gopkg.in/mgo.v2/bson"
)

type AdminConfig struct {
	EnablePath   bool
	EnableDomain bool
	Path         string
	Domain       string
	Admins       map[string]string
}

type CaptchaData struct {
	ImgPath  string
	Question string
	Answer   string
}

type Counter struct {
	Name string
	Seq  int64
}

type PageInfo struct {
	Page     int
	HasPrev  bool
	PrevPage int
	HasNext  bool
	NextPage int
	MaxPage  int
}

type Post struct {
	Id           bson.ObjectId "_id"
	ThreadId     bson.ObjectId
	Num          int32
	Author       User
	Content      string
	Date         int64
	LastModified int64
	ContentType  string
}

type PostInfo struct {
	Data            Post
	StrDate         string
	StrLastModified string
	IsDeleted       bool
	Editable        bool
	Modified        bool
	IsAnon          bool
}

type SiteInfo struct {
	Title              string
	Secret             string
	HomeThreadsNum     int
	HomeTagsNum        int
	PostsPerPage       int
	ThreadsPerPage     int
	TagResultsPerPage  int
	TagsPerPage        int
	MaxPostLength      int
	Footer             []string
	PostFooter         string
	FullVersionDomain  string
	LightVersionDomain string
	UseProxy           bool
	ProxyDomain        string
	ProxyRoot          string
}

type Tag struct {
	Name       string
	Posts      int64
	LastUpdate int64
	LastThread bson.ObjectId
}

type TagData struct {
	Name          string
	URLName       string
	LastUpdate    int64
	LastThread    ThreadInfo
	LastIndex     int64
	StrLastUpdate string
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

type ThreadInfo struct {
	Thread      Thread
	LastPost    PostInfo
	LastMessage int
	Page        int
}

type User struct {
	Nickname       string
	Tripcode       string
	HiddenTripcode bool
	Ip             string
}
