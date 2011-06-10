package main

import (
	"launchpad.net/gobson/bson"
	"launchpad.net/mgo"
	"os"
	"time"
)

type Post struct {
	Id			bson.ObjectId "_id/c"
	Title		string
	Content		string
	Timestamp	int64
}

type PostModel struct {
	c			*mgo.Collection
}

func PostModelInit(c mgo.Collection) *PostModel {
	p := new(PostModel)
	p.c = &c
	return p
}

func (post *PostModel) FrontPage() []map[string]string {
	var result *Post
	results := []map[string]string{}
	err := post.c.Find(nil).Sort(bson.M{"timestamp":-1}).Limit(10).For(&result, func() os.Error {
		t := time.SecondsToLocalTime(result.Timestamp)
		results = append(results, map[string]string {"Title":result.Title, "Content":result.Content, "Date":t.Format("2006 Jan 02 15:04"), "Id": objectIdHex(result.Id.String())})
		return nil
	})
	if err != nil {
		panic(err)
	}
	
	return results
}

func (post *PostModel) Get(postId string) *Post {
	var result *Post
	err := post.c.Find(bson.M{"_id": bson.ObjectIdHex(postId)}).One(&result)
	if err != nil {
		panic(err)
	}
	
	return result
}

func (post *PostModel) GetNextId(postId string) (string, bool) {
	var next *Post
	result := post.Get(postId)
	err := post.c.Find(bson.M{"timestamp": bson.M{"$gt":result.Timestamp}}).Sort(bson.M{"timestamp":1}).One(&next)
	if err != nil && err != mgo.NotFound {
		panic(err)
	}
	
	if err == mgo.NotFound {
		return "", false
	}
	return objectIdHex(next.Id.String()), true
}

func (post *PostModel) GetLastId(postId string) (string, bool) {
	var last *Post
	result := post.Get(postId)
	err := post.c.Find(bson.M{"timestamp": bson.M{"$lt":result.Timestamp}}).Sort(bson.M{"timestamp":-1}).One(&last)
	if err != nil && err != mgo.NotFound {
		panic(err)
	}
	
	if err == mgo.NotFound {
		return "", false
	}
	return objectIdHex(last.Id.String()), true
}

func (post *PostModel) PostListing() []map[string]string {
	var result *Post
	results := []map[string]string{}
	err := post.c.Find(nil).Sort(bson.M{"timestamp":-1}).For(&result, func() os.Error {
		t := time.SecondsToLocalTime(result.Timestamp)
		results = append(results, map[string]string {"Title":result.Title, "Date":t.Format("2006 Jan 02 15:04"), "Id": objectIdHex(result.Id.String())})
		return nil
	})
	if err != nil {
		panic(err)
	}
	return results
}

func (post *PostModel) Create(title string, content string) {
	t := time.LocalTime()
	err := post.c.Insert(&Post{"", title, content, t.Seconds()})
	if err != nil {
		panic(err)
	}
}

func (post *PostModel) Update(title string, content string, postId string) {
	result := post.Get(postId)
	result.Title = title
	result.Content = content
	err := post.c.Update(bson.M{"_id":bson.ObjectIdHex(postId)},result)
	if err != nil {
		panic(err)
	}
}

func (post *PostModel) Delete(postId string) {
	err := post.c.Remove(bson.M{"_id":bson.ObjectIdHex(postId)})
	if err != nil {
		panic(err)
	}
}