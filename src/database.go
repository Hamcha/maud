package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"strings"
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
		Id:          pid,
		ThreadId:    tid,
		Author:      user,
		Content:     content,
		Date:        now,
		ContentType: "bbcode",
	}

	err := database.C("threads").Insert(thread)
	if err != nil {
		return "", err
	}
	err = database.C("posts").Insert(post)

	// Increase tag popularity
	for i := range tags {
		DBIncTag(tags[i], tid)
	}

	return thread.ShortUrl, err
}

func DBReplyThread(thread *Thread, user User, content string) (int, error) {
	post := Post{
		Id:          bson.NewObjectId(),
		ThreadId:    thread.Id,
		Author:      user,
		Content:     content,
		Date:        time.Now().UTC().Unix(),
		ContentType: "bbcode",
	}

	err := database.C("posts").Insert(post)
	if err != nil {
		return 0, err
	}

	err = database.C("threads").UpdateId(thread.Id, bson.M{
		"$set": bson.M{
			"lastreply": post.Id,
			"lrdate":    post.Date,
		},
		"$inc": bson.M{
			"messages": 1,
		},
	})

	// Increase tag popularity
	for i := range thread.Tags {
		DBIncTag(thread.Tags[i], thread.Id)
	}

	return int(thread.Messages), err
}

func DBGetThreadList(tag string, limit, offset int) ([]Thread, error) {
	var filterByTag bson.M
	if tag != "" {
		if idx := strings.IndexRune(tag, ','); idx > 0 {
			// tag1,tag2,... means 'union'
			tags := strings.Split(tag, ",")
			tagsRgx := make([]bson.RegEx, len(tags))
			for i, _ := range tags {
				tagsRgx[i] = bson.RegEx{strings.TrimSpace(tags[i]), "i"}
			}
			filterByTag = bson.M{"tags": bson.M{"$in": tagsRgx}}
		} else {
			// single tag
			filterByTag = bson.M{"tags": bson.RegEx{tag, "i"}}
		}
	} else {
		filterByTag = nil
	}
	query := database.C("threads").Find(filterByTag).Sort("-lrdate")
	if offset > 0 {
		query = query.Skip(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	var threads []Thread
	err := query.All(&threads)
	return threads, err
}

func DBGetThread(surl string) (Thread, error) {
	var thread Thread
	err := database.C("threads").Find(bson.M{"shorturl": surl}).One(&thread)
	return thread, err
}

func DBGetThreadById(id bson.ObjectId) (Thread, error) {
	var thread Thread
	err := database.C("threads").FindId(id).One(&thread)
	return thread, err
}

func DBGetPost(id bson.ObjectId) (Post, error) {
	var post Post
	err := database.C("posts").FindId(id).One(&post)
	return post, err
}

func DBGetPosts(thread *Thread, limit, offset int) ([]Post, error) {
	query := database.C("posts").Find(bson.M{"threadid": thread.Id}).Sort("date")
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

func DBPostCount(thread *Thread) (int, error) {
	return database.C("posts").Find(bson.M{"threadid": thread.Id}).Count()
}

type ByThreads []Tag

func (b ByThreads) Len() int           { return len(b) }
func (b ByThreads) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByThreads) Less(i, j int) bool { return b[i].Posts < b[j].Posts }

func DBGetPopularTags(limit, offset int) ([]Tag, error) {
	var result []Tag
	query := database.C("tags").Find(nil).Sort("-lastupdate")
	if offset > 0 {
		query = query.Skip(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.All(&result)
	sort.Sort(sort.Reverse(ByThreads(result)))
	return result, err
}

func DBNextId(name string) (int64, error) {
	inc := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"seq": 1}},
		Upsert:    true,
		ReturnNew: true,
	}
	var doc Counter
	_, err := database.C("counters").Find(bson.M{"name": name}).Apply(inc, &doc)
	return doc.Seq, err
}

func DBIncTag(name string, lastThread bson.ObjectId) error {
	inc := mgo.Change{
		Update: bson.M{
			"$inc": bson.M{"posts": 1},
			"$set": bson.M{
				"lastupdate": time.Now().UTC().Unix(),
				"lastthread": lastThread,
			},
		},
		Upsert:    true,
		ReturnNew: true,
	}
	var doc Tag
	_, err := database.C("tags").Find(bson.M{"name": name}).Apply(inc, &doc)
	return err
}

func DBEditPost(id bson.ObjectId, newContent string) error {
	err := database.C("posts").UpdateId(id, bson.M{
		"$set": bson.M{
			"lastmodified": time.Now().UTC().Unix(),
			"content":      newContent,
		},
	})
	return err
}
