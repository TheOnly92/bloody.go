package main

import (
	"launchpad.net/gobson/bson"
	"launchpad.net/mgo"
	"./mustache"
	"os"
	"time"
)

func index() string {
	c := mSession.DB(dbname).C("posts")
	var result *Post
	results := []map[string]string{}
	err := c.Find(nil).Sort(bson.M{"timestamp":-1}).Limit(10).For(&result, func() os.Error {
		t := time.SecondsToLocalTime(result.Timestamp)
		results = append(results, map[string]string {"Title":result.Title, "Content":result.Content, "Date":t.Format("2006 Jan 02 15:04"), "Id": objectIdHex(result.Id.String())})
		return nil
	})
	if err != nil {
		panic(err)
	}
	
	output := mustache.RenderFile("templates/post.mustache", map[string][]map[string]string {"posts":results})
	return layout.Render(map[string]string {"Body": output})
}

func readPost(postId string) string {
	c := mSession.DB(dbname).C("posts")
	var result *Post
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(postId)}).One(&result)
	if err != nil {
		panic(err)
	}
	viewVars := make(map[string]string)
	viewVars["Title"] = result.Title
	viewVars["Content"] = result.Content
	viewVars["Date"] = time.SecondsToLocalTime(result.Timestamp).Format("2006 Jan 02 15:04")
	viewVars["Id"] = objectIdHex(result.Id.String())
	
	var next *Post
	var last *Post
	err = c.Find(bson.M{"timestamp": bson.M{"$gt":result.Timestamp}}).Sort(bson.M{"timestamp":1}).One(&next)
	if err != nil && err != mgo.NotFound {
		panic(err)
	}
	err = c.Find(bson.M{"timestamp": bson.M{"$lt":result.Timestamp}}).Sort(bson.M{"timestamp":-1}).One(&last)
	if err != nil && err != mgo.NotFound {
		panic(err)
	}
	if next != nil {
		viewVars["Next"] = objectIdHex(next.Id.String())
	}
	if last != nil {
		viewVars["Last"] = objectIdHex(last.Id.String())
	}
	
	output := mustache.RenderFile("templates/view-post.mustache", viewVars)
	return layout.Render(map[string]interface{} {"Body": output, "Title": map[string]string {"Name": result.Title}})
}

func listPosts() string {
	c := mSession.DB(dbname).C("posts")
	var result *Post
	results := []map[string]string{}
	err := c.Find(nil).Sort(bson.M{"timestamp":-1}).For(&result, func() os.Error {
		t := time.SecondsToLocalTime(result.Timestamp)
		results = append(results, map[string]string {"Title":result.Title, "Date":t.Format("2006 Jan 02 15:04"), "Id": objectIdHex(result.Id.String())})
		return nil
	})
	if err != nil {
		panic(err)
	}
	
	output := mustache.RenderFile("templates/post-listing.mustache", map[string]interface{} {"Posts": results})
	return layout.Render(map[string]interface{} {"Body": output, "Title": map[string]string {"Name": "Post Listing"}})
}