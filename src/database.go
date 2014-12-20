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

// DBReplyThread appends a reply to the thread `thread`.
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

// DBGetThreadList returns a slice of Threads according to the given
// `tag` string. If `tag` is a single word, return all threads with
// a tag matching it (i.e. thread.tag ~ /tag/i); else, return all
// threads with at least 1 tag matching at least 1 of the words in
// the `tag` string.
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

// DBGetThread returns the thread identified by the shorturl `surl`.
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

// DBGetPosts fetches the posts of a thread. If `limit` is > 0, fetch
// up to `limit` posts. If `offset` > 0, start from `offset`-th post.
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

// DBGetPopularTags returns a slice of Tags (up to `limit`, starting
// from `offset`-th) ordered by "popularity". Popularity is greater
// for tags whose threads have been updated more recently.
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

// DBIncTag updates the tag named `name` by increasing its number of
// posts by 1 and changing lastthread and lastupdate to the correct values.
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

// DBEditPost updates the post with id `id` by changing its content to
// newContent and its lastmodified date to the current date.
func DBEditPost(id bson.ObjectId, newContent string) error {
	err := database.C("posts").UpdateId(id, bson.M{
		"$set": bson.M{
			"lastmodified": time.Now().UTC().Unix(),
			"content":      newContent,
		},
	})
	return err
}

// DBGetMatchingTags returns a slice of Tags matching the given word.
// If word given is "", the behaviour is the same as DBGetPopularTags.
func DBGetMatchingTags(word string, limit, offset int) ([]Tag, error) {
	if len(word) < 1 {
		return DBGetPopularTags(limit, offset)
	}
	matching := bson.M{"name": bson.RegEx{word, "i"}}
	query := database.C("tags").Find(matching).Sort("-lrdate")
	if offset > 0 {
		query = query.Skip(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	var tags []Tag
	err := query.All(&tags)
	return tags, err
}
