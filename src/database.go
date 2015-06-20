package main

import (
	"./data"
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Database struct {
	session  *mgo.Session
	database *mgo.Database
}

func InitDatabase(servers, dbname string) Database {
	var db Database
	var err error
	db.session, err = mgo.Dial(servers)
	if err != nil {
		panic(err)
	}
	db.database = db.session.DB(dbname)
	return db
}

func (db Database) Close() {
	db.session.Close()
}

func (db Database) NewThread(user data.User, title, content string, tags []string) (string, error) {
	tid := bson.NewObjectId()
	pid := bson.NewObjectId()
	now := time.Now().UTC().Unix()

	thread := data.Thread{
		Id:         tid,
		ShortUrl:   generateURL(db, "thread"),
		Title:      title,
		Author:     user,
		Tags:       tags,
		Date:       now,
		Messages:   1,
		ThreadPost: pid,
		LastReply:  pid,
		LRDate:     now,
	}

	post := data.Post{
		Id:          pid,
		ThreadId:    tid,
		Author:      user,
		Content:     content,
		Date:        now,
		ContentType: "bbcode",
	}

	err := db.database.C("threads").Insert(thread)
	if err != nil {
		return "", err
	}
	err = db.database.C("posts").Insert(post)

	// Increase tag popularity
	for i := range tags {
		db.IncTag(tags[i], tid)
	}

	return thread.ShortUrl, err
}

// ReplyThread appends a reply to the thread `thread` and increases the popularity
// of all the thread tags.
// Returns the number of posts in the thread (after this was inserted) and any error
// which may have happened during the transaction.
func (db Database) ReplyThread(thread *data.Thread, user data.User, content string) (int, error) {
	post := data.Post{
		Id:          bson.NewObjectId(),
		ThreadId:    thread.Id,
		Author:      user,
		Content:     content,
		Date:        time.Now().UTC().Unix(),
		ContentType: "bbcode",
	}

	err := db.database.C("posts").Insert(post)
	if err != nil {
		return 0, err
	}

	err = db.database.C("threads").UpdateId(thread.Id, bson.M{
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
		db.IncTag(thread.Tags[i], thread.Id)
	}

	return int(thread.Messages), err
}

// GetThreadList returns a slice of Threads according to the given
// `tag` string. If `tag` is a single word, return all threads with
// a tag matching it (i.e. thread.tag ~ /tag/i); else, return all
// threads with at least 1 tag matching at least 1 of the words in
// the `tag` string. Separator is "#"
func (db Database) GetThreadList(tag string, limit, offset int, hThreads, hTags []string) (threads []data.Thread, err error) {
	var filterByTag bson.M
	tag, err = url.QueryUnescape(tag)
	tag = html.UnescapeString(tag)
	if err != nil {
		return
	}
	if tag != "" {
		if idx := strings.Index(tag, "#"); idx > -1 {
			// tag1,tag2,... means 'union'
			tags := strings.Split(tag, "#")
			var tagsRgx []bson.RegEx
			for i := range tags {
				t := strings.TrimSpace(tags[i])
				if len(t) == 0 {
					continue
				}
				tagsRgx = append(tagsRgx, bson.RegEx{t, "i"})
				// limit query to this number of tags
				if len(tagsRgx) == 10 {
					break
				}
			}
			filterByTag = bson.M{"tags": bson.M{"$in": tagsRgx}}
		} else {
			// single tag:
			filterByTag = bson.M{"tags": bson.RegEx{tag, "i"}}
		}
	} else {
		filterByTag = nil
	}
	// intersect query with hidden threads/tags
	var cond bson.M
	if hThreads != nil && len(hThreads) > 0 {
		cond = bson.M{"shorturl": bson.M{"$nin": hThreads}}
	}
	if hTags != nil && len(hTags) > 0 {
		cond2 := bson.M{"tags": bson.M{"$nin": hTags}}
		if cond != nil {
			cond = bson.M{"$and": []bson.M{cond, cond2}}
		} else {
			cond = cond2
		}
	}
	if cond != nil {
		if filterByTag != nil {
			filterByTag = bson.M{"$and": []bson.M{filterByTag, cond}}
		} else {
			filterByTag = cond
		}
	}
	query := db.database.C("threads").Find(filterByTag).Sort("-lrdate")
	if offset > 0 {
		query = query.Skip(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.All(&threads)
	return
}

// GetThread returns the thread identified by the shorturl `surl`.
func (db Database) GetThread(surl string) (data.Thread, error) {
	var thread data.Thread
	err := db.database.C("threads").Find(bson.M{"shorturl": surl}).One(&thread)
	return thread, err
}

// GetThreads is like GetThread but accepts a string of shorturls.
func (db Database) GetThreads(surl []string, limit, offset int) ([]data.Thread, error) {
	var threads []data.Thread
	query := db.database.C("threads").Find(bson.M{"shorturl": bson.M{"$in": surl}})
	if offset > 0 {
		query = query.Skip(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.All(&threads)
	return threads, err
}

// GetTags returns the list of specified tags
func (db Database) GetTags(tagnames []string, limit, offset int) ([]data.Tag, error) {
	var tags []data.Tag
	query := db.database.C("tags").Find(bson.M{"name": bson.M{"$in": tagnames}})
	if offset > 0 {
		query = query.Skip(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.All(&tags)
	return tags, err
}

func (db Database) GetThreadById(id bson.ObjectId) (data.Thread, error) {
	var thread data.Thread
	err := db.database.C("threads").FindId(id).One(&thread)
	return thread, err
}

func (db Database) GetPost(id bson.ObjectId) (data.Post, error) {
	var post data.Post
	err := db.database.C("posts").FindId(id).One(&post)
	return post, err
}

// GetPosts fetches the posts of a thread. If `limit` is > 0, fetch
// up to `limit` posts. If `offset` > 0, start from `offset`-th post.
func (db Database) GetPosts(thread *data.Thread, limit, offset int) ([]data.Post, error) {
	query := db.database.C("posts").Find(bson.M{"threadid": thread.Id}).Sort("date")
	if offset > 0 {
		query = query.Skip(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	var posts []data.Post
	err := query.All(&posts)

	return posts, err
}

func (db Database) PostCount(thread *data.Thread) (int, error) {
	return db.database.C("posts").Find(bson.M{"threadid": thread.Id}).Count()
}

type ByThreads []data.Tag

func (b ByThreads) Len() int           { return len(b) }
func (b ByThreads) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByThreads) Less(i, j int) bool { return b[i].Posts < b[j].Posts }

// GetPopularTags returns a slice of Tags (up to `limit`, starting
// from `offset`-th) ordered by "popularity". Popularity is greater
// for tags whose threads have been updated more recently.
func (db Database) GetPopularTags(limit, offset int, filter []string) ([]data.Tag, error) {
	var result []data.Tag
	var cond bson.M
	if filter != nil && len(filter) > 0 {
		var condParts []bson.M
		for i := range filter {
			condParts = append(condParts, bson.M{"name": filter[i]})
		}
		cond = bson.M{"$nor": condParts}
	} else {
		cond = nil
	}
	query := db.database.C("tags").Find(cond).Sort("-lastupdate")
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

func (db Database) NextId(name string) (int64, error) {
	inc := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"seq": 1}},
		Upsert:    true,
		ReturnNew: true,
	}
	var doc data.Counter
	_, err := db.database.C("counters").Find(bson.M{"name": name}).Apply(inc, &doc)
	return doc.Seq, err
}

// IncTag updates the tag named `name` by increasing its number of
// posts by 1 and changing lastthread and lastupdate to the correct values.
func (db Database) IncTag(name string, lastThread bson.ObjectId) error {
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
	var doc data.Tag
	_, err := db.database.C("tags").Find(bson.M{"name": name}).Apply(inc, &doc)
	return err
}

// DecTag updates the tag named `name` by decreasing its number of
// posts by 1. LastThread remains the old one.
func (db Database) DecTag(name string) error {
	inc := mgo.Change{
		Update: bson.M{
			"$inc": bson.M{"posts": -1},
		},
		ReturnNew: true,
	}
	var doc data.Tag
	_, err := db.database.C("tags").Find(bson.M{"name": name}).Apply(inc, &doc)
	// remove tag if not referred by any thread.
	if doc.Posts < 1 {
		err = db.database.C("tags").Remove(bson.M{"name": name})
	}
	return err
}

// EditPost updates the post with id `id` by changing its content to
// newContent and its lastmodified date to the current date.
func (db Database) EditPost(id bson.ObjectId, newContent, newContentType string) error {
	err := db.database.C("posts").UpdateId(id, bson.M{
		"$set": bson.M{
			"lastmodified": time.Now().UTC().Unix(),
			"content":      newContent,
			"contenttype":  newContentType,
		},
	})
	return err
}

// DeletePost updates the post with id `id` by changing its content to
// "deleted" or "admin-deleted"
func (db Database) DeletePost(id bson.ObjectId, admin bool) error {
	ctype := "deleted"
	if admin {
		ctype = "admin-deleted"
	}
	err := db.database.C("posts").UpdateId(id, bson.M{
		"$set": bson.M{
			"lastmodified": time.Now().UTC().Unix(),
			"contenttype":  ctype,
		},
	})
	return err
}

// PurgePost deletes a post from the database (IRREVERSIBLE)
// As this action requires many queries, this is considered
// a quite expensive operation.
// Returns an errors as well as a bool indicating if the thread
// has been purged as well
func (db Database) PurgePost(post data.Post) (error, bool) {
	// Get thread
	thread, err := db.GetThreadById(post.ThreadId)
	if err != nil {
		return err, false
	}
	if thread.ThreadPost == post.Id {
		return db.PurgeThread(&thread), true
	}
	if thread.LastReply == post.Id {
		// Get the new last post
		posts, err := db.GetPosts(&thread, 1, int(thread.Messages)-2)
		if err != nil {
			return err, false
		}
		if len(posts) < 1 {
			return errors.New("Something is wrong on this thread.."), false
		}
		// Decrement message count and update last post
		err = db.database.C("threads").UpdateId(thread.Id, bson.M{
			"$set": bson.M{
				"lastreply": posts[0].Id,
				"lrdate":    posts[0].Date,
			},
			"$inc": bson.M{
				"messages": -1,
			},
		})
		if err != nil {
			return err, false
		}
	} else {
		// Just decrement the post count
		err = db.database.C("threads").UpdateId(thread.Id, bson.M{
			"$inc": bson.M{
				"messages": -1,
			},
		})
		if err != nil {
			return err, false
		}
	}

	// Remove post from database
	err = db.database.C("posts").RemoveId(post.Id)
	return err, false
}

// PurgeThread deletes an entire thread from the database (IRREVERSIBLE)
// This also deletes all the posts inside it.
func (db Database) PurgeThread(thread *data.Thread) error {
	// Firstly, remove thread from database
	err := db.database.C("threads").RemoveId(thread.Id)
	if err != nil {
		return err
	}
	// Purge all posts from the thread
	err = db.database.C("posts").Remove(bson.M{
		"threadid": thread.Id,
	})
	if err != nil {
		return err
	}
	// Decrement all tags
	for i := range thread.Tags {
		err = db.DecTag(thread.Tags[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// SetThreadTags changes the tags of the thread with id `id` to `newTags`
// and returns an error, or nil if no error occurred
func (db Database) SetThreadTags(id bson.ObjectId, newTags []string) error {
	err := db.database.C("threads").UpdateId(id, bson.M{
		"$set": bson.M{"tags": newTags},
	})
	return err
}

// GetMatchingTags returns a slice of Tags matching the given word.
// If word given is "", the behaviour is the same as DBGetPopularTags.
func (db Database) GetMatchingTags(word string, limit, offset int, filter []string) ([]data.Tag, error) {
	if len(word) < 1 {
		return db.GetPopularTags(limit, offset, filter)
	}
	matching := bson.M{"name": bson.RegEx{word, "i"}}
	if filter != nil {
		cond := bson.M{"tags": bson.M{"$nin": filter}}
		matching = bson.M{"$and": []bson.M{matching, cond}}
	}
	query := db.database.C("tags").Find(matching).Sort("-lrdate")
	if offset > 0 {
		query = query.Skip(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	var tags []data.Tag
	err := query.All(&tags)
	return tags, err
}
